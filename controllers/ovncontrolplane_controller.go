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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=ovncontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=ovncontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=ovncontrolplanes/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

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

	pkiResources := ovn.PKIResources(instance)
	if err := template.EnsureResources(ctx, r.Client, pkiResources, log); err != nil {
		return ctrl.Result{}, err
	}

	ovsdbNorthSvc, err := r.reconcileOVSDB(ctx, instance, ovn.OVSDBNorth, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	ovsdbSouthSvc, err := r.reconcileOVSDB(ctx, instance, ovn.OVSDBSouth, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	ovsdbConnConfigMap := ovn.OVSDBConnectionConfigMap(instance, ovsdbNorthSvc, ovsdbSouthSvc)
	if err := template.EnsureConfigMap(ctx, r.Client, ovsdbConnConfigMap, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNorthd(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileOVSNode(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileController(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OVNControlPlaneReconciler) reconcileNorthd(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, log logr.Logger) error {
	deploy := ovn.NorthdDeployment(instance)
	controllerutil.SetControllerReference(instance, deploy, r.Scheme)
	if err := template.EnsureDeployment(ctx, r.Client, deploy, log); err != nil {
		return err
	}

	return nil
}

func (r *OVNControlPlaneReconciler) reconcileOVSDB(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, component string, log logr.Logger) (*corev1.Service, error) {
	svc := ovn.OVSDBService(instance, component)
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
		return nil, err
	}

	sts := ovn.OVSDBStatefulSet(instance, component)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return nil, err
	}

	return svc, nil
}

func (r *OVNControlPlaneReconciler) reconcileController(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, log logr.Logger) error {
	ds := ovn.ControllerDaemonSet(instance)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *OVNControlPlaneReconciler) reconcileOVSNode(ctx context.Context, instance *openstackv1beta1.OVNControlPlane, log logr.Logger) error {
	ds := ovn.OVSNodeDaemonSet(instance)
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
