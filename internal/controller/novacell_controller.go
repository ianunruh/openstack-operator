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
	mariadbdatabase "github.com/ianunruh/openstack-operator/pkg/mariadb/database"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	novacell "github.com/ianunruh/openstack-operator/pkg/nova/cell"
	"github.com/ianunruh/openstack-operator/pkg/nova/computeset"
	rabbitmquser "github.com/ianunruh/openstack-operator/pkg/rabbitmq/user"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaCellReconciler reconciles a NovaCell object
type NovaCellReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacells,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacells/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacells/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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

	reporter := novacell.NewReporter(instance, r.Client, r.Recorder)

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

	deps := template.NewConditionWaiter(r.Scheme, log)

	db := nova.CellDatabase(cluster.Name, instance.Spec.Name, instance.Namespace, instance.Spec.Database)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadbdatabase.Ensure(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	}
	mariadbdatabase.AddReadyCheck(deps, db)

	brokerUser := nova.BrokerUser(instance.Name, instance.Namespace, instance.Spec.Broker)
	controllerutil.SetControllerReference(instance, brokerUser, r.Scheme)
	if err := rabbitmquser.Ensure(ctx, r.Client, brokerUser, log); err != nil {
		return ctrl.Result{}, err
	}
	rabbitmquser.AddReadyCheck(deps, brokerUser)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
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

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_API_DATABASE__CONNECTION", cluster.Spec.APIDatabase.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneSecret)...)
	env = append(env, keystone.ClientEnv("OS_NEUTRON__", cluster.Spec.Neutron.Secret)...)
	env = append(env, keystone.ClientEnv("OS_PLACEMENT__", cluster.Spec.Placement.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-nova", cm.Name, nil),
	}

	jobs := template.NewJobRunner(ctx, r.Client, instance, log)
	jobs.Add(&instance.Status.DBSyncJobHash,
		novacell.DBSyncJob(instance, env, volumes, cluster.Spec.API.Image))
	if result, err := jobs.Run(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileConductor(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMetadata(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNoVNCProxy(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileComputeSets(ctx, instance, log); err != nil {
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

func (r *NovaCellReconciler) reconcileComputeSets(ctx context.Context, instance *openstackv1beta1.NovaCell, log logr.Logger) error {
	for name, spec := range instance.Spec.Compute {
		set := computeset.New(instance, name, spec)
		controllerutil.SetControllerReference(instance, set, r.Scheme)
		if err := computeset.Ensure(ctx, r.Client, set, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *NovaCellReconciler) reconcileConductor(ctx context.Context, instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := nova.ConductorService(instance.Name, instance.Namespace)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	sts := nova.ConductorStatefulSet(instance.Name, instance.Namespace, instance.Spec.Conductor, env, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	return nil
}

func (r *NovaCellReconciler) reconcileMetadata(ctx context.Context, instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	env = append(env,
		// TODO make this configurable
		template.SecretEnvVar("OS_NEUTRON__METADATA_PROXY_SHARED_SECRET", "nova", "metadata-proxy-secret"))

	svc := nova.MetadataService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	deploy := nova.MetadataDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	return nil
}

func (r *NovaCellReconciler) reconcileNoVNCProxy(ctx context.Context, instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	svc := nova.NoVNCProxyService(instance)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return err
	}

	if instance.Spec.NoVNCProxy.Ingress == nil {
		// TODO ensure ingress does not exist
	} else {
		ingress := nova.NoVNCProxyIngress(instance)
		controllerutil.SetControllerReference(instance, ingress, r.Scheme)
		if err := template.EnsureIngress(ctx, r.Client, ingress, log); err != nil {
			return err
		}
	}

	deploy := nova.NoVNCProxyDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

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
