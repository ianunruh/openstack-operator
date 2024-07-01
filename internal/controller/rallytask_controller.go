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

	batchv1 "k8s.io/api/batch/v1"
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
	"github.com/ianunruh/openstack-operator/pkg/rally"
	rallytask "github.com/ianunruh/openstack-operator/pkg/rally/task"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// RallyTaskReconciler reconciles a RallyTask object
type RallyTaskReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=rallytasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=rallytasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=rallytasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RallyTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.RallyTask{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.Status.CompletionTime != nil {
		return ctrl.Result{}, nil
	}

	reporter := rallytask.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	cluster := &openstackv1beta1.Rally{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rally",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := reporter.Pending(ctx, "Rally %s not found", cluster.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	rally.AddReadyCheck(deps, cluster)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	env := []corev1.EnvVar{
		template.SecretEnvVar("OS_DATABASE__CONNECTION", cluster.Spec.Database.Secret, "connection"),
		template.EnvVar("RALLY_TASK_PATH", instance.Spec.Path),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-rally", cluster.Name, nil),
		// template.PersistentVolume("data", template.Combine(cluster.Name, "data")),
	}

	job := rallytask.RunnerJob(instance, cluster, env, volumes)
	controllerutil.SetControllerReference(instance, job, r.Scheme)
	if err := template.CreateJob(ctx, r.Client, job, log); err != nil {
		return ctrl.Result{}, err
	} else if job.Status.CompletionTime == nil {
		if err := reporter.Pending(ctx, "Waiting on Job %s condition Complete", job.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	instance.Status.CompletionTime = job.Status.CompletionTime

	if err := reporter.Completed(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RallyTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.RallyTask{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
