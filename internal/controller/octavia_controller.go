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

package controller

import (
	"context"

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
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystonesvc "github.com/ianunruh/openstack-operator/pkg/keystone/service"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	"github.com/ianunruh/openstack-operator/pkg/octavia"
	"github.com/ianunruh/openstack-operator/pkg/octavia/amphora"
	rabbitmquser "github.com/ianunruh/openstack-operator/pkg/rabbitmq/user"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// OctaviaReconciler reconciles a Octavia object
type OctaviaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=octavias,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=octavias/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=octavias/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *OctaviaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Octavia{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ks := &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(ks), ks); err != nil {
		return ctrl.Result{}, err
	}

	pkiResources := amphora.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	deps := template.NewConditionWaiter(log)

	db := octavia.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	mariadbdatabase.AddReadyCheck(deps, db)

	brokerUser := octavia.BrokerUser(instance)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmquser.Ensure(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	}
	rabbitmquser.AddReadyCheck(deps, brokerUser)

	keystoneSvc := octavia.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystonesvc.Ensure(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}
	keystonesvc.AddReadyCheck(deps, keystoneSvc)

	keystoneUser := octavia.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, keystoneUser)

	if err := octavia.EnsureKeystoneRoles(ctx, instance, r.Client); err != nil {
		return ctrl.Result{}, err
	}

	if result := deps.Wait(); !result.IsZero() {
		return result, nil
	}

	amphoraSecret := amphora.Secret(instance)
	if err := template.CreateSecret(ctx, r.Client, amphoraSecret, log); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := amphora.Bootstrap(ctx, instance, r.Client, log); err != nil || !result.IsZero() {
		return result, err
	}

	cm := octavia.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_HEALTH_MANAGER__HEARTBEAT_KEY", amphoraSecret.Name, "heartbeat-key"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
		template.ConfigMapEnvVar("OS_OVN__OVN_NB_CONNECTION", "ovn-ovsdb", "OVN_NB_CONNECTION"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)
	env = append(env, keystone.ClientEnv("OS_SERVICE_AUTH__", keystoneUser.Spec.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-octavia", cm.Name, nil),
		template.HostPathVolume("host-var-run-octavia", "/var/run/octavia"),
	}

	jobs := template.NewJobRunner(ctx, r.Client, log)
	jobs.Add(&instance.Status.DBSyncJobHash, octavia.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(instance); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileAPI(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDriverAgent(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileHealthManager(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileHousekeeping(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileWorker(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	// TODO wait for deploys to be ready then mark status

	return ctrl.Result{}, nil
}

func (r *OctaviaReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	svc := octavia.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := octavia.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := octavia.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *OctaviaReconciler) reconcileDriverAgent(ctx context.Context, instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	if !instance.Spec.OVN.Enabled {
		// TODO ensure deployment does not exist
		return nil
	}

	deploy := octavia.DriverAgentDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *OctaviaReconciler) reconcileHealthManager(ctx context.Context, instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	ds := octavia.HealthManagerDaemonSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *OctaviaReconciler) reconcileHousekeeping(ctx context.Context, instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	deploy := octavia.HousekeepingDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *OctaviaReconciler) reconcileWorker(ctx context.Context, instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	deploy := octavia.WorkerDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OctaviaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Octavia{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}).
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
