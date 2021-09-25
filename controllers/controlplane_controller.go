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
	"github.com/ianunruh/openstack-operator/pkg/barbican"
	"github.com/ianunruh/openstack-operator/pkg/cinder"
	"github.com/ianunruh/openstack-operator/pkg/controlplane"
	"github.com/ianunruh/openstack-operator/pkg/glance"
	"github.com/ianunruh/openstack-operator/pkg/heat"
	"github.com/ianunruh/openstack-operator/pkg/horizon"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	keystoneuser "github.com/ianunruh/openstack-operator/pkg/keystone/user"
	"github.com/ianunruh/openstack-operator/pkg/magnum"
	"github.com/ianunruh/openstack-operator/pkg/manila"
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/memcached"
	"github.com/ianunruh/openstack-operator/pkg/neutron"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/octavia"
	"github.com/ianunruh/openstack-operator/pkg/ovn"
	"github.com/ianunruh/openstack-operator/pkg/placement"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/rally"
)

// ControlPlaneReconciler reconciles a ControlPlane object
type ControlPlaneReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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

	ovnControlPlane := controlplane.OVNControlPlane(instance)
	controllerutil.SetControllerReference(instance, ovnControlPlane, r.Scheme)
	if err := ovn.EnsureControlPlane(ctx, r.Client, ovnControlPlane, log); err != nil {
		return ctrl.Result{}, err
	}

	cache := controlplane.Cache(instance)
	controllerutil.SetControllerReference(instance, cache, r.Scheme)
	if err := memcached.EnsureCluster(ctx, r.Client, cache, log); err != nil {
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

	if len(instance.Spec.Cinder.Backends) > 0 {
		volume := controlplane.Cinder(instance)
		controllerutil.SetControllerReference(instance, volume, r.Scheme)
		if err := cinder.EnsureCinder(ctx, r.Client, volume, log); err != nil {
			return ctrl.Result{}, err
		}
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

	dashboard := controlplane.Horizon(instance)
	controllerutil.SetControllerReference(instance, dashboard, r.Scheme)
	if err := horizon.EnsureHorizon(ctx, r.Client, dashboard, log); err != nil {
		return ctrl.Result{}, err
	}

	if instance.Spec.Barbican.Image != "" {
		keyManager := controlplane.Barbican(instance)
		controllerutil.SetControllerReference(instance, keyManager, r.Scheme)
		if err := barbican.EnsureBarbican(ctx, r.Client, keyManager, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Spec.Heat.Image != "" {
		orchestration := controlplane.Heat(instance)
		controllerutil.SetControllerReference(instance, orchestration, r.Scheme)
		if err := heat.EnsureHeat(ctx, r.Client, orchestration, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Spec.Magnum.Image != "" {
		containerInfra := controlplane.Magnum(instance)
		controllerutil.SetControllerReference(instance, containerInfra, r.Scheme)
		if err := magnum.EnsureMagnum(ctx, r.Client, containerInfra, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Spec.Octavia.Image != "" {
		loadBalancer := controlplane.Octavia(instance)
		controllerutil.SetControllerReference(instance, loadBalancer, r.Scheme)
		if err := octavia.EnsureOctavia(ctx, r.Client, loadBalancer, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Spec.Manila.Image != "" {
		sfs := controlplane.Manila(instance)
		controllerutil.SetControllerReference(instance, sfs, r.Scheme)
		if err := manila.EnsureManila(ctx, r.Client, sfs, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Spec.Rally.Image != "" {
		benchmark := controlplane.Rally(instance)
		controllerutil.SetControllerReference(instance, benchmark, r.Scheme)
		if err := rally.EnsureRally(ctx, r.Client, benchmark, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	demoUser := controlplane.DemoKeystoneUser(instance)
	controllerutil.SetControllerReference(instance, demoUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, demoUser, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.ControlPlane{}).
		Owns(&openstackv1beta1.Cinder{}).
		Owns(&openstackv1beta1.Glance{}).
		Owns(&openstackv1beta1.Horizon{}).
		Owns(&openstackv1beta1.Keystone{}).
		Owns(&openstackv1beta1.MariaDB{}).
		Owns(&openstackv1beta1.Memcached{}).
		Owns(&openstackv1beta1.Neutron{}).
		Owns(&openstackv1beta1.Nova{}).
		Owns(&openstackv1beta1.Placement{}).
		Owns(&openstackv1beta1.RabbitMQ{}).
		Complete(r)
}
