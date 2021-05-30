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
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// KeystoneUserReconciler reconciles a KeystoneUser object
type KeystoneUserReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=keystoneusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=keystoneusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=keystoneusers/finalizers,verbs=update

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

	cluster := &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		return ctrl.Result{}, err
	} else if !cluster.Status.Ready {
		log.Info("Waiting on Keystone to be available", "name", cluster.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	secret := keystone.UserSecret(instance)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	if err := template.CreateSecret(ctx, r.Client, secret, log); err != nil {
		return ctrl.Result{}, err
	}

	job := keystone.UserJob(instance, cluster.Spec.Image, cluster.Name)
	controllerutil.SetControllerReference(instance, job, r.Scheme)

	runner := template.NewJobRunner(ctx, r.Client, job, log)
	runner.CheckStatus(func(hash string) bool {
		return instance.Status.SetupJobHash == hash
	})
	runner.UpdateStatus(func(hash string) {
		instance.Status.Ready = true
		instance.Status.SetupJobHash = hash
	})
	return runner.Run(instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeystoneUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.KeystoneUser{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
