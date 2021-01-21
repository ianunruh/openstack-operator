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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/magnum"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// MagnumReconciler reconciles a Magnum object
type MagnumReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=magnums,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=magnums/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=magnums/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *MagnumReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Magnum{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	db := magnum.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	} else if !db.Status.Ready {
		log.Info("Waiting on database to be available", "name", db.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	brokerUser := magnum.BrokerUser(instance)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmq.EnsureUser(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !brokerUser.Status.Ready {
		log.Info("Waiting on broker to be available", "name", brokerUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	keystoneSvc := magnum.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystone.EnsureService(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO domain admin user
	keystoneUser := magnum.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}

	keystoneStackUser := magnum.KeystoneStackUser(instance)
	controllerutil.SetControllerReference(instance, keystoneStackUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneStackUser, log); err != nil {
		return ctrl.Result{}, err
	}

	if !keystoneUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else if !keystoneStackUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneStackUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	cm := magnum.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	envVars := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_KEYSTONE_AUTH__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_TRUST__TRUSTEE_DOMAIN_ADMIN_PASSWORD", keystoneStackUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-magnum", cm.Name, nil),
	}

	jobs := []*batchv1.Job{
		magnum.DBSyncJob(instance, envVars, volumes),
		// magnum.BootstrapJob(instance),
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

	if err := r.reconcileAPI(ctx, instance, envVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConductor(ctx, instance, envVars, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *MagnumReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Magnum, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := magnum.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := magnum.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := magnum.APIDeployment(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *MagnumReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.Magnum, envVars []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := magnum.ConductorService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := magnum.ConductorStatefulSet(instance, envVars, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MagnumReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Magnum{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&netv1.Ingress{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&openstackv1beta1.KeystoneService{}).
		Owns(&openstackv1beta1.KeystoneUser{}).
		Owns(&openstackv1beta1.MariaDBDatabase{}).
		Owns(&openstackv1beta1.RabbitMQUser{}).
		Complete(r)
}
