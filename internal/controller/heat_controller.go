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
	"fmt"
	"time"

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
	"github.com/ianunruh/openstack-operator/pkg/heat"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystonesvc "github.com/ianunruh/openstack-operator/pkg/keystone/service"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	rabbitmquser "github.com/ianunruh/openstack-operator/pkg/rabbitmq/user"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// HeatReconciler reconciles a Heat object
type HeatReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=heats,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=heats/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=heats/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HeatReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Heat{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := heat.NewReporter(instance, r.Client, r.Recorder)

	ks := &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(ks), ks); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := reporter.Pending(ctx, "Keystone %s not found", ks.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	deps := template.NewConditionWaiter(r.Scheme, log)

	db := heat.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	mariadbdatabase.AddReadyCheck(deps, db)

	brokerUser := heat.BrokerUser(instance)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmquser.Ensure(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	}
	rabbitmquser.AddReadyCheck(deps, brokerUser)

	keystoneServices := heat.KeystoneServices(instance)
	for _, keystoneSvc := range keystoneServices {
		controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
		if err := keystonesvc.Ensure(ctx, r.Client, keystoneSvc, log); err != nil {
			return ctrl.Result{}, err
		}
		keystonesvc.AddReadyCheck(deps, keystoneSvc)
	}

	keystoneUser := heat.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, keystoneUser)

	// TODO need to create heat_stack_user role

	keystoneStackUser := heat.KeystoneStackUser(instance)
	controllerutil.SetControllerReference(instance, keystoneStackUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneStackUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, keystoneStackUser)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	// TODO if disabled, clean up resources
	pkiResources := heat.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	cm := heat.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("REQUESTS_CA_BUNDLE", "/etc/ssl/certs/ca-certificates.crt"),
		template.EnvVar("OS_CLIENTS_KEYSTONE__AUTH_URI", fmt.Sprintf("https://%s", ks.Spec.API.Ingress.Host)),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DEFAULT__STACK_DOMAIN_ADMIN_PASSWORD", keystoneStackUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
		template.SecretEnvVar("OS_TRUSTEE__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)
	env = append(env, keystone.ClientEnv("OS_TRUSTEE__", keystoneUser.Spec.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-heat", cm.Name, nil),
	}

	jobs := template.NewJobRunner(ctx, r.Client, instance, log)
	jobs.Add(&instance.Status.DBSyncJobHash, heat.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileAPI(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileCFN(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileEngine(ctx, instance, env, volumes, deps, log); err != nil {
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

func (r *HeatReconciler) reconcileAPI(ctx context.Context, instance *openstackv1beta1.Heat, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := heat.APIService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := heat.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := heat.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	return nil
}

func (r *HeatReconciler) reconcileCFN(ctx context.Context, instance *openstackv1beta1.Heat, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := heat.CFNService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.CFN.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := heat.CFNIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := heat.CFNDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	return nil
}

func (r *HeatReconciler) reconcileEngine(ctx context.Context, instance *openstackv1beta1.Heat, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := heat.EngineService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := heat.EngineStatefulSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HeatReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Heat{}).
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
