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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/glance"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystonesvc "github.com/ianunruh/openstack-operator/pkg/keystone/service"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// GlanceReconciler reconciles a Glance object
type GlanceReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=glances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=glances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=glances/finalizers,verbs=update
//+kubebuilder:rbac:groups=ceph.rook.io,resources=cephblockpools,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=ceph.rook.io,resources=cephclients,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GlanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Glance{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := glance.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	db := glance.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	mariadbdatabase.AddReadyCheck(deps, db)

	keystoneSvc := glance.KeystoneService(instance)
	controllerutil.SetControllerReference(instance, keystoneSvc, r.Scheme)
	if err := keystonesvc.Ensure(ctx, r.Client, keystoneSvc, log); err != nil {
		return ctrl.Result{}, err
	}
	keystonesvc.AddReadyCheck(deps, keystoneSvc)

	keystoneUser := glance.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, keystoneUser)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	// TODO if disabled, clean up resources
	pkiResources := glance.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	cm := glance.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("REQUESTS_CA_BUNDLE", "/etc/ssl/certs/ca-certificates.crt"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneUser.Spec.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-glance", instance.Name, nil),
	}

	for _, pvc := range glance.PersistentVolumeClaims(instance) {
		controllerutil.SetControllerReference(instance, pvc, r.Scheme)
		if err := template.EnsurePersistentVolumeClaim(ctx, r.Client, pvc, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileRookCeph(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	jobs := template.NewJobRunner(ctx, r.Client, instance, log)
	jobs.Add(&instance.Status.DBSyncJobHash, glance.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	service := glance.APIService(instance)
	controllerutil.SetControllerReference(instance, service, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, service, log); err != nil {
		return ctrl.Result{}, err
	}

	if instance.Spec.API.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := glance.APIIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	deploy := glance.APIDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return ctrl.Result{}, err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := reporter.Running(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GlanceReconciler) reconcileRookCeph(ctx context.Context, instance *openstackv1beta1.Glance, log logr.Logger) error {
	resources := glance.RookCephResources(instance)
	for _, resource := range resources {
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return err
		}
	}

	secrets, err := glance.RookCephSecrets(ctx, r.Client, instance)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		controllerutil.SetControllerReference(instance, secret, r.Scheme)
		if err := template.EnsureSecret(ctx, r.Client, secret, log); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Glance{}).
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
