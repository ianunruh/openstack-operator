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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	novaflavor "github.com/ianunruh/openstack-operator/pkg/nova/flavor"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaFlavorReconciler reconciles a NovaFlavor object
type NovaFlavorReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novaflavors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novaflavors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novaflavors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NovaFlavorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaFlavor{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := novaflavor.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	cluster := &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if controllerutil.RemoveFinalizer(instance, template.Finalizer) {
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		if err := reporter.Pending(ctx, "Nova %s not found", cluster.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	nova.AddReadyCheck(deps, cluster)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova-keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(svcUser), svcUser); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if controllerutil.RemoveFinalizer(instance, template.Finalizer) {
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		if err := reporter.Pending(ctx, "Secret %s not found", svcUser.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
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

		if err := novaflavor.Delete(instance, compute, log); err != nil {
			if err := reporter.DeleteError(ctx, "Error deleting Nova flavor: %w", err); err != nil {
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

	if err := novaflavor.Reconcile(ctx, r.Client, instance, compute, log); err != nil {
		if err := reporter.Error(ctx, "Error reconciling Nova flavor: %v", err); err != nil {
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
func (r *NovaFlavorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaFlavor{}).
		Complete(r)
}
