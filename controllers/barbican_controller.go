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
	"github.com/ianunruh/openstack-operator/pkg/barbican"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// BarbicanReconciler reconciles a Barbican object
type BarbicanReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=barbicans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=barbicans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=barbicans/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BarbicanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Barbican{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	db := barbican.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	} else if !db.Status.Ready {
		log.Info("Waiting on database to be available", "name", db.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	brokerUser := barbican.BrokerUser(instance)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmq.EnsureUser(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !brokerUser.Status.Ready {
		log.Info("Waiting on broker to be available", "name", brokerUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	keystoneSvc := barbican.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystone.EnsureService(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}

	keystoneUser := barbican.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	} else if !keystoneUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	secret := barbican.Secret(instance)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	if err := template.CreateSecret(ctx, r.Client, secret, log); err != nil {
		return ctrl.Result{}, err
	}

	cm := barbican.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("OS_DEFAULT__HOST_HREF", fmt.Sprintf("https://%s", instance.Spec.API.Ingress.Host)),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DEFAULT__SQL_CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_SIMPLE_CRYPTO_PLUGIN__KEK", instance.Name, "kek"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-barbican", cm.Name, nil),
	}

	jobs := template.NewJobRunner(ctx, r.Client, log)
	jobs.Add(&instance.Status.DBSyncJobHash, barbican.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(instance); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileAPI(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileWorker(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *BarbicanReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Barbican, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := barbican.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := barbican.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := barbican.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *BarbicanReconciler) reconcileWorker(ctx context.Context, instance *openstackv1beta1.Barbican, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	deploy := barbican.WorkerDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BarbicanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Barbican{}).
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
