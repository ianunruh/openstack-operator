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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaReconciler reconciles a Nova object
type NovaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=nova,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=nova/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=nova/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NovaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Nova{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	databases := []*openstackv1beta1.MariaDBDatabase{
		nova.APIDatabase(instance),
		nova.CellDatabase(instance.Name, "cell0", instance.Namespace, instance.Spec.CellDatabase),
	}
	for _, db := range databases {
		controllerutil.SetControllerReference(instance, db, r.Scheme)
		if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
			return ctrl.Result{}, err
		} else if !db.Status.Ready {
			log.Info("Waiting on database to be available", "name", db.Name)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	brokerUser := nova.BrokerUser(instance.Name, instance.Namespace, instance.Spec.Broker)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmq.EnsureUser(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !brokerUser.Status.Ready {
		log.Info("Waiting on broker to be available", "name", brokerUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	keystoneSvc := nova.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystone.EnsureService(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}

	keystoneUser := nova.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !keystoneUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	cm := nova.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	secret := nova.Secret(instance)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	if err := template.CreateSecret(ctx, r.Client, secret, log); err != nil {
		return ctrl.Result{}, err
	}

	envVars := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_PLACEMENT__PASSWORD", "placement-keystone", "OS_PASSWORD"),
		template.SecretEnvVar("OS_NEUTRON__PASSWORD", "neutron-keystone", "OS_PASSWORD"),
	}

	dbEnvVars := []corev1.EnvVar{
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_API_DATABASE__CONNECTION", instance.Spec.APIDatabase.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.CellDatabase.Secret, "connection"),
	}

	fullEnvVars := append(envVars, dbEnvVars...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-nova", cm.Name, nil),
	}

	jobs := []*batchv1.Job{
		nova.DBSyncJob(instance, fullEnvVars, volumes),
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

	if err := r.reconcileAPI(ctx, instance, fullEnvVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	for _, cellSpec := range instance.Spec.Cells {
		cell := nova.Cell(instance, cellSpec)
		controllerutil.SetControllerReference(instance, cell, r.Scheme)
		if err := nova.EnsureCell(ctx, r.Client, cell, log); err != nil {
			return ctrl.Result{}, err
		} else if !cell.Status.Ready {
			log.Info("Waiting on NovaCell to be ready", "name", cell.Name)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	if err := r.reconcileConductor(ctx, instance, fullEnvVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileScheduler(ctx, instance, fullEnvVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileLibvirtd(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileCompute(ctx, instance, envVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *NovaReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Nova, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := nova.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	ingress := nova.APIIngress(instance)
	controllerutil.SetControllerReference(instance, ingress, r.Scheme)
	if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
		return err
	}

	deploy := nova.APIDeployment(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.Nova, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := nova.ConductorService(instance.Name, instance.Namespace)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.ConductorStatefulSet(instance.Name, instance.Namespace, instance.Spec.Conductor, envVars, volumes, instance.Spec.Image)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaReconciler) reconcileScheduler(ctx context.Context, instance *openstackv1beta1.Nova, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := nova.SchedulerService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.SchedulerStatefulSet(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaReconciler) reconcileLibvirtd(ctx context.Context, instance *openstackv1beta1.Nova, log logr.Logger) error {
	cm := nova.LibvirtdConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return err
	}
	configHash := template.AppliedHash(cm)

	ds := nova.LibvirtdDaemonSet(instance, configHash)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaReconciler) reconcileCompute(ctx context.Context, instance *openstackv1beta1.Nova, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	// TODO make this configurable
	cell := instance.Spec.Cells[0]

	extraEnvVars := []corev1.EnvVar{
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", cell.Broker.Secret, "connection"),
		template.EnvVar("OS_VNC__NOVNCPROXY_BASE_URL", fmt.Sprintf("https://%s/vnc_auto.html", cell.NoVNCProxy.Ingress.Host)),
	}

	envVars = append(envVars, extraEnvVars...)

	ds := nova.ComputeDaemonSet(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Nova{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&netv1.Ingress{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&openstackv1beta1.KeystoneService{}).
		Owns(&openstackv1beta1.KeystoneUser{}).
		Owns(&openstackv1beta1.MariaDBDatabase{}).
		Owns(&openstackv1beta1.NovaCell{}).
		Owns(&openstackv1beta1.RabbitMQUser{}).
		Complete(r)
}
