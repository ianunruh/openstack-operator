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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// KeystoneUserReconciler reconciles a KeystoneUser object
type KeystoneUserReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=keystoneusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=keystoneusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=keystoneusers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KeystoneUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.KeystoneUser{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := keystoneuser.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	cluster := &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
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
		if err := reporter.Pending(ctx, "Keystone %s not found", instance.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	keystone.AddReadyCheck(deps, cluster)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
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

	identity, err := keystone.NewIdentityServiceClient(ctx, svcUser)
	if err != nil {
		return ctrl.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		// resource marked for deletion
		if !controllerutil.ContainsFinalizer(instance, template.Finalizer) {
			return ctrl.Result{}, nil
		}

		if err := keystoneuser.Delete(instance, identity, log); err != nil {
			if err := reporter.DeleteError(ctx, "Error deleting Keystone user: %w", err); err != nil {
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

	var currentPassword string
	currentSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.Secret,
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(currentSecret), currentSecret); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	} else {
		currentPassword = keystoneuser.PasswordFromSecret(currentSecret)
	}

	secret := keystoneuser.Secret(instance, cluster, currentPassword)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	if err := template.EnsureSecret(ctx, r.Client, secret, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := keystoneuser.Reconcile(instance, secret, identity, log); err != nil {
		if err := reporter.Error(ctx, "Error reconciling Keystone user: %v", err); err != nil {
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
func (r *KeystoneUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.KeystoneUser{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
