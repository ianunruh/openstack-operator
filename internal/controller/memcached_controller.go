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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/memcached"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=memcacheds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=memcacheds/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Memcached{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := memcached.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(log)

	if err := r.reconcileServices(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	sts := memcached.ClusterStatefulSet(instance)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return ctrl.Result{}, err
	}
	template.AddStatefulSetReadyCheck(deps, sts)

	if err := r.reconcileServiceMonitor(ctx, instance, log); err != nil {
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

func (r *MemcachedReconciler) reconcileServices(ctx context.Context, instance *openstackv1beta1.Memcached, log logr.Logger) error {
	services := []*corev1.Service{
		memcached.ClusterService(instance),
		memcached.ClusterHeadlessService(instance),
	}

	for _, svc := range services {
		controllerutil.SetControllerReference(instance, svc, r.Scheme)
		err := template.EnsureService(ctx, r.Client, svc, log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MemcachedReconciler) reconcileServiceMonitor(ctx context.Context, instance *openstackv1beta1.Memcached, log logr.Logger) error {
	promSpec := instance.Spec.Prometheus
	if promSpec == nil || !promSpec.ServiceMonitor {
		// TODO ensure service monitor does not exist
		return nil
	}

	svcMonitor := memcached.ClusterServiceMonitor(instance)
	controllerutil.SetControllerReference(instance, svcMonitor, r.Scheme)
	if err := template.EnsureResource(ctx, r.Client, svcMonitor, log); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Memcached{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
