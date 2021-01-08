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
	"github.com/ianunruh/openstack-operator/pkg/neutron"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NeutronReconciler reconciles a Neutron object
type NeutronReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=neutrons,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=neutrons/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=neutrons/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NeutronReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Neutron{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	db := neutron.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	// TODO wait for database to be ready

	brokerUser := neutron.BrokerUser(instance)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmq.EnsureUser(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	}
	// TODO wait for broker to be ready

	keystoneSvc := neutron.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystone.EnsureService(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}

	keystoneUser := neutron.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}

	cm := neutron.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	envVars := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_NOVA__PASSWORD", "nova-keystone", "OS_PASSWORD"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-neutron", cm.Name, nil),
	}

	jobs := []*batchv1.Job{
		neutron.DBSyncJob(instance, envVars, volumes),
		// neutron.BootstrapJob(instance),
	}
	for _, job := range jobs {
		controllerutil.SetControllerReference(instance, job, r.Scheme)
		if err := template.CreateJob(ctx, r.Client, job, log); err != nil {
			return ctrl.Result{}, err
		}
		// TODO wait for job to finish
	}

	if err := r.reconcileServer(ctx, instance, envVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *NeutronReconciler) reconcileServer(ctx context.Context, instance *openstackv1beta1.Neutron, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := neutron.ServerService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	ingress := neutron.ServerIngress(instance)
	controllerutil.SetControllerReference(instance, ingress, r.Scheme)
	if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
		return err
	}

	deploy := neutron.ServerDeployment(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NeutronReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Neutron{}).
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
