/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaCellReconciler reconciles a NovaCell object
type NovaCellReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacells,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacells/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacells/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NovaCellReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaCell{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster := &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
	}
	err = r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	db := nova.CellDatabase(cluster.Name, instance.Spec.Name, instance.Namespace, instance.Spec.Database)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	} else if !db.Status.Ready {
		log.Info("Waiting on database to be available", "name", db.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	brokerUser := nova.BrokerUser(instance.Name, instance.Namespace, instance.Spec.Broker)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmq.EnsureUser(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !brokerUser.Status.Ready {
		log.Info("Waiting on broker to be available", "name", brokerUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cm), cm); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	keystoneSecret := template.Combine(cluster.Name, "keystone")

	envVars := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_API_DATABASE__CONNECTION", cluster.Spec.APIDatabase.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneSecret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_PLACEMENT__PASSWORD", "placement-keystone", "OS_PASSWORD"),
		template.SecretEnvVar("OS_NEUTRON__PASSWORD", "neutron-keystone", "OS_PASSWORD"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-nova", cm.Name, nil),
	}

	jobs := []*batchv1.Job{
		nova.CellDBSyncJob(instance, envVars, volumes, cluster.Spec.Image),
		// nova.BootstrapJob(instance),
	}
	for _, job := range jobs {
		controllerutil.SetControllerReference(instance, job, r.Scheme)
		if err := template.CreateJob(ctx, r.Client, job, log); err != nil {
			return ctrl.Result{}, err
		} else if job.Status.CompletionTime == nil {
			log.Info("Waiting on job completion", "name", job.Name)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	if err := r.reconcileConductor(ctx, instance, envVars, volumes, cluster.Spec.Image, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMetadata(ctx, instance, envVars, volumes, cluster.Spec.Image, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNoVNCProxy(ctx, instance, envVars, volumes, cluster.Spec.Image, log); err != nil {
		return ctrl.Result{}, err
	}

	if !instance.Status.Ready {
		instance.Status.Ready = true
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NovaCellReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.NovaCell, envVars []corev1.EnvVar, volumes []corev1.Volume, containerImage string, log logr.Logger) error {
	svc := nova.ConductorService(instance.Name, instance.Namespace)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.ConductorStatefulSet(instance.Name, instance.Namespace, instance.Spec.Conductor, envVars, volumes, containerImage)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaCellReconciler) reconcileMetadata(ctx context.Context, instance *openstackv1beta1.NovaCell, envVars []corev1.EnvVar, volumes []corev1.Volume, containerImage string, log logr.Logger) error {
	svc := nova.MetadataService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	deploy := nova.MetadataDeployment(instance, envVars, volumes, containerImage)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaCellReconciler) reconcileNoVNCProxy(ctx context.Context, instance *openstackv1beta1.NovaCell, envVars []corev1.EnvVar, volumes []corev1.Volume, containerImage string, log logr.Logger) error {
	svc := nova.NoVNCProxyService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	ingress := nova.NoVNCProxyIngress(instance)
	controllerutil.SetControllerReference(instance, ingress, r.Scheme)
	if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
		return err
	}

	deploy := nova.NoVNCProxyDeployment(instance, envVars, volumes, containerImage)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaCellReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaCell{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&netv1.Ingress{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&openstackv1beta1.KeystoneService{}).
		Owns(&openstackv1beta1.KeystoneUser{}).
		Owns(&openstackv1beta1.MariaDBDatabase{}).
		Complete(r)
}
