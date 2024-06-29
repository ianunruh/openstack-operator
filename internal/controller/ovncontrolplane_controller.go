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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/ovn"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// OVNControlPlaneReconciler reconciles a OVNControlPlane object
type OVNControlPlaneReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=ovncontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=ovncontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=ovncontrolplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *OVNControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.OVNControlPlane{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := ovn.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	pkiResources := ovn.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	cm := ovn.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-ovn", cm.Name, nil),
	}

	if err := r.reconcileAllOVSDB(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := r.reconcileNorthd(ctx, instance, env, volumes, deps, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileOVSNode(ctx, instance, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileController(ctx, instance, env, volumes, log); err != nil {
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

func (r *OVNControlPlaneReconciler) reconcileNorthd(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	deploy := ovn.NorthdDeployment(instance, env, volumes)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}
	template.AddDeploymentReadyCheck(deps, deploy)

	return nil
}

func (r *OVNControlPlaneReconciler) reconcileAllOVSDB(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) error {
	northSvc, err := r.reconcileOVSDB(ctx, instance, ovn.OVSDBNorth, env, volumes, deps, log)
	if err != nil {
		return err
	}

	southSvc, err := r.reconcileOVSDB(ctx, instance, ovn.OVSDBSouth, env, volumes, deps, log)
	if err != nil {
		return err
	}

	ovsdbConnConfigMap := ovn.OVSDBConnectionConfigMap(instance, northSvc, southSvc)
	controllerutil.SetControllerReference(instance, ovsdbConnConfigMap, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, ovsdbConnConfigMap, log); err != nil {
		return err
	}

	return nil
}

func (r *OVNControlPlaneReconciler) reconcileOVSDB(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, component string, env []corev1.EnvVar, volumes []corev1.Volume, deps *template.ConditionWaiter, log logr.Logger) (*corev1.Service, error) {
	svc := ovn.OVSDBService(instance, component)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return nil, err
	}

	sts := ovn.OVSDBStatefulSet(instance, component, env, volumes)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return nil, err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	return svc, nil
}

func (r *OVNControlPlaneReconciler) reconcileController(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	ds := ovn.ControllerDaemonSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *OVNControlPlaneReconciler) reconcileOVSNode(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	ds := ovn.OVSNodeDaemonSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OVNControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.OVNControlPlane{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
