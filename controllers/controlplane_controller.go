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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/controlplane"
	"github.com/ianunruh/openstack-operator/pkg/glance"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/neutron"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/placement"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
)

// ControlPlaneReconciler reconciles a ControlPlane object
type ControlPlaneReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=controlplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=controlplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=controlplanes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControlPlane object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.ControlPlane{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	database := controlplane.Database(instance)
	controllerutil.SetControllerReference(instance, database, r.Scheme)
	if err := mariadb.EnsureCluster(ctx, r.Client, database, log); err != nil {
		return ctrl.Result{}, err
	}

	broker := controlplane.Broker(instance)
	controllerutil.SetControllerReference(instance, broker, r.Scheme)
	if err := rabbitmq.EnsureCluster(ctx, r.Client, broker, log); err != nil {
		return ctrl.Result{}, err
	}

	identity := controlplane.Keystone(instance)
	controllerutil.SetControllerReference(instance, identity, r.Scheme)
	if err := keystone.EnsureKeystone(ctx, r.Client, identity, log); err != nil {
		return ctrl.Result{}, err
	}

	image := controlplane.Glance(instance)
	controllerutil.SetControllerReference(instance, image, r.Scheme)
	if err := glance.EnsureGlance(ctx, r.Client, image, log); err != nil {
		return ctrl.Result{}, err
	}

	pm := controlplane.Placement(instance)
	controllerutil.SetControllerReference(instance, pm, r.Scheme)
	if err := placement.EnsurePlacement(ctx, r.Client, pm, log); err != nil {
		return ctrl.Result{}, err
	}

	compute := controlplane.Nova(instance)
	controllerutil.SetControllerReference(instance, compute, r.Scheme)
	if err := nova.EnsureNova(ctx, r.Client, compute, log); err != nil {
		return ctrl.Result{}, err
	}

	network := controlplane.Neutron(instance)
	controllerutil.SetControllerReference(instance, network, r.Scheme)
	if err := neutron.EnsureNeutron(ctx, r.Client, network, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.ControlPlane{}).
		Complete(r)
}
