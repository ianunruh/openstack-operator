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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova/computeset"
)

// NovaComputeSetReconciler reconciles a NovaComputeSet object
type NovaComputeSetReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NovaComputeSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaComputeSet{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := computeset.Reconcile(ctx, r.Client, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaComputeSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	nodeRequestsFn := handler.MapFunc(func(object client.Object) []reconcile.Request {
		c := mgr.GetClient()

		var sets openstackv1beta1.NovaComputeSetList
		if err := c.List(context.Background(), &sets); err != nil {
			r.Log.Error(err, "Failed to list NovaComputeSet")
			return nil
		}

		var requests []reconcile.Request
		for _, instance := range sets.Items {
			// TODO limit to compute sets that matched the changed node
			name := client.ObjectKey{
				Namespace: instance.Namespace,
				Name:      instance.Name,
			}
			requests = append(requests, reconcile.Request{NamespacedName: name})
		}

		return requests
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaComputeSet{}).
		Watches(&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(nodeRequestsFn)).
		Complete(r)
}
