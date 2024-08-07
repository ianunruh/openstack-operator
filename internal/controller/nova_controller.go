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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystonesvc "github.com/ianunruh/openstack-operator/pkg/keystone/service"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	novacell "github.com/ianunruh/openstack-operator/pkg/nova/cell"
	novaflavor "github.com/ianunruh/openstack-operator/pkg/nova/flavor"
	rabbitmquser "github.com/ianunruh/openstack-operator/pkg/rabbitmq/user"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaReconciler reconciles a Nova object
type NovaReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novas/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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

	reporter := nova.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	databases := []*openstackv1beta1.MariaDBDatabase{
		nova.APIDatabase(instance),
		nova.CellDatabase(instance.Name, "cell0", instance.Namespace, instance.Spec.CellDatabase),
	}
	for _, db := range databases {
		controllerutil.SetControllerReference(instance, db, r.Scheme)
		if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
			return ctrl.Result{}, err
		}
		mariadbdatabase.AddReadyCheck(deps, db)
	}

	brokerUser := nova.BrokerUser(instance.Name, instance.Namespace, instance.Spec.Broker)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmquser.Ensure(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	}
	rabbitmquser.AddReadyCheck(deps, brokerUser)

	keystoneSvc := nova.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystonesvc.Ensure(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}
	keystonesvc.AddReadyCheck(deps, keystoneSvc)

	keystoneUser := nova.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, keystoneUser)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	cinder := &openstackv1beta1.Cinder{
		// TODO make this configurable
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cinder",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cinder), cinder); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		cinder = nil
	}

	// TODO if disabled, clean up resources
	pkiResources := nova.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	cm := nova.ConfigMap(instance, cinder)
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

	computeSSHSecret, err := nova.ComputeSSHKeypairSecret(instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	controllerutil.SetControllerReference(instance, computeSSHSecret, r.Scheme)
	if err := template.CreateSecret(ctx, r.Client, computeSSHSecret, log); err != nil {
		return ctrl.Result{}, err
	}

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("REQUESTS_CA_BUNDLE", "/etc/ssl/certs/ca-certificates.crt"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)
	env = append(env, keystone.ClientEnv("OS_NEUTRON__", instance.Spec.Neutron.Secret)...)
	env = append(env, keystone.ClientEnv("OS_PLACEMENT__", instance.Spec.Placement.Secret)...)

	dbEnvVars := []corev1.EnvVar{
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_API_DATABASE__CONNECTION", instance.Spec.APIDatabase.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.CellDatabase.Secret, "connection"),
	}

	fullEnvVars := append(env, dbEnvVars...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-nova", cm.Name, nil),
	}

	jobs := template.NewJobRunner(ctx, r.Client, instance, log)
	jobs.Add(&instance.Status.DBSyncJobHash, nova.DBSyncJob(instance, fullEnvVars, volumes))
	if result, err := jobs.Run(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileAPI(ctx, instance, fullEnvVars, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	for _, cellSpec := range instance.Spec.Cells {
		cell := nova.Cell(instance, cellSpec)
		controllerutil.SetControllerReference(instance, cell, r.Scheme)
		if err := novacell.Ensure(ctx, r.Client, cell, log); err != nil {
			return ctrl.Result{}, err
		}
		novacell.AddReadyCheck(deps, cell)
	}

	// wait for cells to be ready
	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileConductor(ctx, instance, fullEnvVars, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileScheduler(ctx, instance, fullEnvVars, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileFlavors(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := reporter.Running(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NovaReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Nova, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := nova.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := nova.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := nova.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	return nil
}

func (r *NovaReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.Nova, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := nova.ConductorService(instance.Name, instance.Namespace)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.ConductorStatefulSet(
		instance.Name,
		instance.Namespace,
		instance.Spec.Conductor,
		instance.Spec.Broker,
		instance.Spec.TLS,
		env,
		volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	return nil
}

func (r *NovaReconciler) reconcileScheduler(ctx context.Context, instance *openstackv1beta1.Nova, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := nova.SchedulerService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.SchedulerStatefulSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	return nil
}

func (r *NovaReconciler) reconcileFlavors(ctx context.Context, instance *openstackv1beta1.Nova, log logr.Logger) error {
	for name, spec := range instance.Spec.Flavors {
		flavor := &openstackv1beta1.NovaFlavor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: instance.Namespace,
			},
			Spec: spec,
		}
		controllerutil.SetControllerReference(instance, flavor, r.Scheme)
		if err := novaflavor.Ensure(ctx, r.Client, flavor, log); err != nil {
			return err
		}
	}

	// TODO cleanup removed/renamed flavors

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
