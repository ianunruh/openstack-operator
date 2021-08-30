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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/rally"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// RallyReconciler reconciles a Rally object
type RallyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rallies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rallies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rallies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rally object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *RallyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.Rally{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	db := rally.Database(instance)
	controllerutil.SetControllerReference(instance, db, r.Scheme)
	if err := mariadb.EnsureDatabase(ctx, r.Client, db, log); err != nil {
		return ctrl.Result{}, err
	} else if !db.Status.Ready {
		log.Info("Waiting on database to be available", "name", db.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	keystoneUser := rally.KeystoneUser(instance)
	controllerutil.SetControllerReference(instance, keystoneUser, r.Scheme)
	if err := keystone.EnsureUser(ctx, r.Client, keystoneUser, log); err != nil {
		return ctrl.Result{}, err
	}

	if !keystoneUser.Status.Ready {
		log.Info("Waiting on Keystone user to be available", "name", keystoneUser.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	cm := rally.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	pvc := rally.PersistentVolumeClaim(instance)
	controllerutil.SetControllerReference(instance, pvc, r.Scheme)
	if err := template.EnsurePersistentVolumeClaim(ctx, r.Client, pvc, log); err != nil {
		return ctrl.Result{}, err
	}

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneUser.Spec.Secret, "OS_PASSWORD"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-rally", cm.Name, nil),
		template.PersistentVolume("data", pvc.Name),
	}

	jobs := template.NewJobRunner(ctx, r.Client, log)
	jobs.Add(&instance.Status.DBSyncJobHash, rally.DBSyncJob(instance, env, volumes))
	if result, err := jobs.Run(instance); err != nil || !result.IsZero() {
		return result, err
	}

	runnerJob := rally.TaskRunnerJob(instance, keystoneUser, env, volumes)
	controllerutil.SetControllerReference(instance, runnerJob, r.Scheme)
	if err := template.CreateJob(ctx, r.Client, runnerJob, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RallyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.Rally{}).
		Complete(r)
}
