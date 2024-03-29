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
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
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
	reporter := keystoneuser.NewReporter(r.Recorder)

	instance := &openstackv1beta1.KeystoneUser{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if keystoneuser.ReadyCondition(instance) == nil {
		reporter.Pending(instance, nil, "KeystoneUserPending", "Waiting for user to be reconciled")
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	deps := template.NewConditionWaiter(log)

	cluster := &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		return ctrl.Result{}, err
	}
	keystone.AddReadyCheck(deps, cluster)

	if result := deps.Wait(); !result.IsZero() {
		return result, nil
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

	// TODO handle this user not existing on deletion, and remove the finalizer anyway
	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(svcUser), svcUser); err != nil {
		return ctrl.Result{}, err
	}

	identity, err := keystone.NewIdentityServiceClient(ctx, svcUser)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := keystoneuser.Reconcile(instance, secret, identity, log); err != nil {
		reporter.Pending(instance, err, "KeystoneUserReconcileError", "Error reconciling Keystone user")
		if statusErr := r.Client.Status().Update(ctx, instance); statusErr != nil {
			err = utilerrors.NewAggregate([]error{statusErr, err})
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	condition := keystoneuser.ReadyCondition(instance)
	if condition.Status == metav1.ConditionFalse {
		reporter.Succeeded(instance)
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
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
