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
	keystonesvc "github.com/ianunruh/openstack-operator/pkg/keystone/service"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	"github.com/ianunruh/openstack-operator/pkg/magnum"
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// MagnumReconciler reconciles a Magnum object
type MagnumReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=magnums,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=magnums/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=magnums/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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

	deps := template.NewConditionWaiter(log)

	db := magnum.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	deps.AddReadyCheck(db, db.Status.Conditions)

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
	if err := keystonesvc.Ensure(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO domain admin user
	keystoneUser := magnum.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}

	keystoneStackUser := magnum.KeystoneStackUser(instance)
	controllerutil.SetControllerReference(instance, keystoneStackUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneStackUser, log); err != nil {
		return ctrl.Result{}, err
	}

	if !keystoneUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else if !keystoneStackUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneStackUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if result := deps.Wait(); !result.IsZero() {
		return result, nil
	}

	cm := magnum.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_KEYSTONE_AUTH__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
		template.SecretEnvVar("OS_TRUST__TRUSTEE_DOMAIN_ADMIN_PASSWORD", keystoneStackUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
	}

	env = append(env, keystone.ClientEnv("OS_KEYSTONE_AUTH__", keystoneUser.Spec.Secret)...)
	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-magnum", cm.Name, nil),
	}

	jobs := template.NewJobRunner(ctx, r.Client, log)
	jobs.Add(&instance.Status.DBSyncJobHash, magnum.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(instance); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileAPI(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConductor(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *MagnumReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Magnum, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
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

	deploy := magnum.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *MagnumReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.Magnum, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := magnum.ConductorService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := magnum.ConductorStatefulSet(instance, env, volumes)
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
