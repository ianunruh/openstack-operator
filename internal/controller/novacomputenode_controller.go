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
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/nova/computenode"
)

// NovaComputeNodeReconciler reconciles a NovaComputeNode object
type NovaComputeNodeReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputenodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputenodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputenodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NovaComputeNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)
	reporter := computenode.NewReporter(r.Recorder)

	instance := &openstackv1beta1.NovaComputeNode{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if computenode.ReadyCondition(instance) == nil {
		reporter.Pending(instance, nil, "ComputeNodePending", "Waiting for compute node to be reconciled")
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// TODO handle this user not existing on deletion, and remove the finalizer anyway
	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova-keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(svcUser), svcUser); err != nil {
		return ctrl.Result{}, err
	}

	compute, err := nova.NewComputeServiceClient(ctx, svcUser)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := computenode.Reconcile(ctx, r.Client, instance, compute, log); err != nil {
		reporter.Pending(instance, err, "ComputeNodeReconcileError", "Error reconciling compute node")
		if statusErr := r.Client.Status().Update(ctx, instance); statusErr != nil {
			err = utilerrors.NewAggregate([]error{statusErr, err})
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	condition := computenode.ReadyCondition(instance)
	if condition.Status == metav1.ConditionFalse {
		reporter.Running(instance)
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaComputeNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaComputeNode{}).
		Complete(r)
}
