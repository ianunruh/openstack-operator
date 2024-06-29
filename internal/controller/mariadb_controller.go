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
	"strings"

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
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// MariaDBReconciler reconciles a MariaDB object
type MariaDBReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=mariadbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=mariadbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=mariadbs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MariaDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.MariaDB{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := mariadb.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	// Create admin secret if not found
	if err := r.reconcileSecret(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	configHashes, err := r.reconcileConfigMap(ctx, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}
	configHash := strings.Join(configHashes, ",")

	if err := r.reconcileServices(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

	sts := mariadb.ClusterStatefulSet(instance, configHash)
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

func (r *MariaDBReconciler) reconcileSecret(ctx context.Context, instance *openstackv1beta1.MariaDB, log logr.Logger) error {
	secret := mariadb.Secret(instance)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	return template.CreateSecret(ctx, r.Client, secret, log)
}

func (r *MariaDBReconciler) reconcileConfigMap(ctx context.Context, instance *openstackv1beta1.MariaDB, log logr.Logger) ([]string, error) {
	cm := mariadb.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return nil, err
	}
	return []string{template.AppliedHash(cm)}, nil
}

func (r *MariaDBReconciler) reconcileServices(ctx context.Context, instance *openstackv1beta1.MariaDB, log logr.Logger) error {
	services := []*corev1.Service{
		mariadb.ClusterService(instance),
		mariadb.ClusterHeadlessService(instance),
	}

	for _, svc := range services {
		controllerutil.SetControllerReference(instance, svc, r.Scheme)
		if err := template.EnsureService(ctx, r.Client, svc, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *MariaDBReconciler) reconcileServiceMonitor(ctx context.Context, instance *openstackv1beta1.MariaDB, log logr.Logger) error {
	promSpec := instance.Spec.Prometheus
	if promSpec == nil || !promSpec.ServiceMonitor {
		// TODO ensure service monitor does not exist
		return nil
	}

	svcMonitor := mariadb.ClusterServiceMonitor(instance)
	controllerutil.SetControllerReference(instance, svcMonitor, r.Scheme)
	if err := template.EnsureResource(ctx, r.Client, svcMonitor, log); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.MariaDB{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
