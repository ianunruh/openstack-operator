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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
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
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// ControlPlaneReconciler reconciles a ControlPlane object
type ControlPlaneReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=controlplanes/finalizers,verbs=update

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

	reporter := controlplane.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	// TODO if disabled, clean up resources
	pkiResources := controlplane.PKIResources(instance)
	for _, resource := range pkiResources {
		controllerutil.SetControllerReference(instance, resource, r.Scheme)
		if err := template.EnsureResource(ctx, r.Client, resource, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	ovnControlPlane := controlplane.OVNControlPlane(instance)
	controllerutil.SetControllerReference(instance, ovnControlPlane, r.Scheme)
	if err := ovn.Ensure(ctx, r.Client, ovnControlPlane, log); err != nil {
		return ctrl.Result{}, err
	}
	ovn.AddReadyCheck(deps, ovnControlPlane)

	if cache := controlplane.Cache(instance); cache != nil {
		controllerutil.SetControllerReference(instance, cache, r.Scheme)
		if err := memcached.Ensure(ctx, r.Client, cache, log); err != nil {
			return ctrl.Result{}, err
		}
		memcached.AddReadyCheck(deps, cache)
	}

	if database := controlplane.Database(instance); database != nil {
		controllerutil.SetControllerReference(instance, database, r.Scheme)
		if err := mariadb.Ensure(ctx, r.Client, database, log); err != nil {
			return ctrl.Result{}, err
		}
		mariadb.AddReadyCheck(deps, database)
	}

	if broker := controlplane.Broker(instance); broker != nil {
		controllerutil.SetControllerReference(instance, broker, r.Scheme)
		if err := rabbitmq.Ensure(ctx, r.Client, broker, log); err != nil {
			return ctrl.Result{}, err
		}
		rabbitmq.AddReadyCheck(deps, broker)
	}

	identity := controlplane.Keystone(instance)
	controllerutil.SetControllerReference(instance, identity, r.Scheme)
	if err := keystone.Ensure(ctx, r.Client, identity, log); err != nil {
		return ctrl.Result{}, err
	}
	keystone.AddReadyCheck(deps, identity)

	image := controlplane.Glance(instance)
	controllerutil.SetControllerReference(instance, image, r.Scheme)
	if err := glance.Ensure(ctx, r.Client, image, log); err != nil {
		return ctrl.Result{}, err
	}
	glance.AddReadyCheck(deps, image)

	pm := controlplane.Placement(instance)
	controllerutil.SetControllerReference(instance, pm, r.Scheme)
	if err := placement.Ensure(ctx, r.Client, pm, log); err != nil {
		return ctrl.Result{}, err
	}
	placement.AddReadyCheck(deps, pm)

	if volume := controlplane.Cinder(instance); volume != nil {
		controllerutil.SetControllerReference(instance, volume, r.Scheme)
		if err := cinder.Ensure(ctx, r.Client, volume, log); err != nil {
			return ctrl.Result{}, err
		}
		cinder.AddReadyCheck(deps, volume)
	}

	compute := controlplane.Nova(instance)
	controllerutil.SetControllerReference(instance, compute, r.Scheme)
	if err := nova.Ensure(ctx, r.Client, compute, log); err != nil {
		return ctrl.Result{}, err
	}
	nova.AddReadyCheck(deps, compute)

	network := controlplane.Neutron(instance)
	controllerutil.SetControllerReference(instance, network, r.Scheme)
	if err := neutron.Ensure(ctx, r.Client, network, log); err != nil {
		return ctrl.Result{}, err
	}
	neutron.AddReadyCheck(deps, network)

	if dashboard := controlplane.Horizon(instance); dashboard != nil {
		controllerutil.SetControllerReference(instance, dashboard, r.Scheme)
		if err := horizon.Ensure(ctx, r.Client, dashboard, log); err != nil {
			return ctrl.Result{}, err
		}
		horizon.AddReadyCheck(deps, dashboard)
	}

	if keyManager := controlplane.Barbican(instance); keyManager != nil {
		controllerutil.SetControllerReference(instance, keyManager, r.Scheme)
		if err := barbican.Ensure(ctx, r.Client, keyManager, log); err != nil {
			return ctrl.Result{}, err
		}
		barbican.AddReadyCheck(deps, keyManager)
	}

	if orchestration := controlplane.Heat(instance); orchestration != nil {
		controllerutil.SetControllerReference(instance, orchestration, r.Scheme)
		if err := heat.Ensure(ctx, r.Client, orchestration, log); err != nil {
			return ctrl.Result{}, err
		}
		heat.AddReadyCheck(deps, orchestration)
	}

	if containerInfra := controlplane.Magnum(instance); containerInfra != nil {
		controllerutil.SetControllerReference(instance, containerInfra, r.Scheme)
		if err := magnum.Ensure(ctx, r.Client, containerInfra, log); err != nil {
			return ctrl.Result{}, err
		}
		magnum.AddReadyCheck(deps, containerInfra)
	}
	if loadBalancer := controlplane.Octavia(instance); loadBalancer != nil {
		controllerutil.SetControllerReference(instance, loadBalancer, r.Scheme)
		if err := octavia.Ensure(ctx, r.Client, loadBalancer, log); err != nil {
			return ctrl.Result{}, err
		}
		octavia.AddReadyCheck(deps, loadBalancer)
	}

	if sfs := controlplane.Manila(instance); sfs != nil {
		controllerutil.SetControllerReference(instance, sfs, r.Scheme)
		if err := manila.Ensure(ctx, r.Client, sfs, log); err != nil {
			return ctrl.Result{}, err
		}
		manila.AddReadyCheck(deps, sfs)
	}

	if benchmark := controlplane.Rally(instance); benchmark != nil {
		controllerutil.SetControllerReference(instance, benchmark, r.Scheme)
		if err := rally.Ensure(ctx, r.Client, benchmark, log); err != nil {
			return ctrl.Result{}, err
		}
		rally.AddReadyCheck(deps, benchmark)
	}

	demoUser := controlplane.DemoKeystoneUser(instance)
	controllerutil.SetControllerReference(instance, demoUser, r.Scheme)
	if err := keystoneuser.Ensure(ctx, r.Client, demoUser, log); err != nil {
		return ctrl.Result{}, err
	}
	keystoneuser.AddReadyCheck(deps, demoUser)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := reporter.Running(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.ControlPlane{}).
		Owns(&openstackv1beta1.Barbican{}).
		Owns(&openstackv1beta1.Cinder{}).
		Owns(&openstackv1beta1.Glance{}).
		Owns(&openstackv1beta1.Heat{}).
		Owns(&openstackv1beta1.Horizon{}).
		Owns(&openstackv1beta1.Keystone{}).
		Owns(&openstackv1beta1.KeystoneUser{}).
		Owns(&openstackv1beta1.Magnum{}).
		Owns(&openstackv1beta1.Manila{}).
		Owns(&openstackv1beta1.MariaDB{}).
		Owns(&openstackv1beta1.Memcached{}).
		Owns(&openstackv1beta1.Neutron{}).
		Owns(&openstackv1beta1.Nova{}).
		Owns(&openstackv1beta1.Octavia{}).
		Owns(&openstackv1beta1.OVNControlPlane{}).
		Owns(&openstackv1beta1.Placement{}).
		Owns(&openstackv1beta1.RabbitMQ{}).
		Owns(&openstackv1beta1.Rally{}).
		Complete(r)
}
