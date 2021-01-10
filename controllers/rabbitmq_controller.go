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
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// RabbitMQReconciler reconciles a RabbitMQ object
type RabbitMQReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rabbitmqs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rabbitmqs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=rabbitmqs/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RabbitMQ object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *RabbitMQReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.RabbitMQ{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcileRBAC(ctx, instance, log); err != nil {
		return ctrl.Result{}, err
	}

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

	sts := rabbitmq.ClusterStatefulSet(instance, configHash)
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	if err := template.EnsureStatefulSet(ctx, r.Client, sts, log); err != nil {
		return ctrl.Result{}, err
	} else if sts.Status.ReadyReplicas == 0 {
		log.Info("Waiting for StatefulSet to be available", "name", sts.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else if !instance.Status.Ready {
		instance.Status.Ready = true
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RabbitMQReconciler) reconcileRBAC(ctx context.Context, instance *openstackv1beta1.RabbitMQ, log logr.Logger) error {
	labels := template.AppLabels(instance.Name, rabbitmq.AppLabel)

	// TODO move some of this to rabbitmq package
	sa := template.GenericServiceAccount(instance.Name, instance.Namespace, labels)
	controllerutil.SetControllerReference(instance, sa, r.Scheme)
	if err := template.EnsureServiceAccount(ctx, r.Client, sa, log); err != nil {
		return err
	}

	role := template.GenericRole(instance.Name, instance.Namespace, labels, []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"endpoints"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"events"},
			Verbs:     []string{"create"},
		},
	})
	controllerutil.SetControllerReference(instance, role, r.Scheme)
	if err := template.EnsureRole(ctx, r.Client, role, log); err != nil {
		return err
	}

	roleBinding := template.GenericRoleBinding(instance.Name, instance.Namespace, labels)
	roleBinding.RoleRef = template.RoleRef(role.Name)
	roleBinding.Subjects = []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: sa.Name},
	}
	controllerutil.SetControllerReference(instance, roleBinding, r.Scheme)
	if err := template.EnsureRoleBinding(ctx, r.Client, roleBinding, log); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQReconciler) reconcileSecret(ctx context.Context, instance *openstackv1beta1.RabbitMQ, log logr.Logger) error {
	secret := rabbitmq.Secret(instance)
	controllerutil.SetControllerReference(instance, secret, r.Scheme)
	return template.CreateSecret(ctx, r.Client, secret, log)
}

func (r *RabbitMQReconciler) reconcileConfigMap(ctx context.Context, instance *openstackv1beta1.RabbitMQ, log logr.Logger) ([]string, error) {
	cm := rabbitmq.ConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	err := template.EnsureConfigMap(ctx, r.Client, cm, log)
	if err != nil {
		return nil, err
	}
	return []string{template.AppliedHash(cm)}, nil
}

func (r *RabbitMQReconciler) reconcileServices(ctx context.Context, instance *openstackv1beta1.RabbitMQ, log logr.Logger) error {
	services := []*corev1.Service{
		rabbitmq.ClusterService(instance),
		rabbitmq.ClusterHeadlessService(instance),
	}

	for _, svc := range services {
		controllerutil.SetControllerReference(instance, svc, r.Scheme)
		err := template.EnsureService(ctx, r.Client, svc, log)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RabbitMQReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.RabbitMQ{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
