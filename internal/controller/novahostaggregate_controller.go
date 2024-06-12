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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/nova/hostaggregate"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaHostAggregateReconciler reconciles a NovaHostAggregate object
type NovaHostAggregateReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novahostaggregates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novahostaggregates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novahostaggregates/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NovaHostAggregateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaHostAggregate{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := hostaggregate.NewReporter(instance, r.Client, r.Recorder)

	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova-keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(svcUser), svcUser); err != nil {
		if errors.IsNotFound(err) {
			controllerutil.RemoveFinalizer(instance, template.Finalizer)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}

	compute, err := nova.NewComputeServiceClient(ctx, svcUser)
	if err != nil {
		return ctrl.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		// resource marked for deletion
		if !controllerutil.ContainsFinalizer(instance, template.Finalizer) {
			return ctrl.Result{}, nil
		}

		if err := hostaggregate.Delete(instance, compute, log); err != nil {
			if err := reporter.DeleteError(ctx, "Error deleting Nova host aggregate: %w", err); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		controllerutil.RemoveFinalizer(instance, template.Finalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(instance, template.Finalizer) {
		controllerutil.AddFinalizer(instance, template.Finalizer)
		if err := r.Client.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := hostaggregate.Reconcile(ctx, r.Client, instance, compute, log); err != nil {
		if err := reporter.Error(ctx, "Error reconciling host aggregate: %v", err); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if err := reporter.Reconciled(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaHostAggregateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	nodeRequestsFn := handler.MapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
		c := mgr.GetClient()

		var aggregates openstackv1beta1.NovaHostAggregateList
		if err := c.List(context.Background(), &aggregates); err != nil {
			r.Log.Error(err, "Failed to list NovaHostAggregate")
			return nil
		}

		var requests []reconcile.Request
		for _, instance := range aggregates.Items {
			if instance.Spec.NodeSelector == nil {
				continue
			}

			name := client.ObjectKey{
				Namespace: instance.Namespace,
				Name:      instance.Name,
			}
			requests = append(requests, reconcile.Request{NamespacedName: name})
		}

		return requests
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaHostAggregate{}).
		Watches(&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(nodeRequestsFn)).
		Complete(r)
}
