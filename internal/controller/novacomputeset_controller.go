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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/nova/computeset"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaComputeSetReconciler reconciles a NovaComputeSet object
type NovaComputeSetReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.ospk8s.com,resources=novacomputesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NovaComputeSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaComputeSet{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reporter := computeset.NewReporter(instance, r.Client, r.Recorder)

	deps := template.NewConditionWaiter(r.Scheme, log)

	cluster := &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := reporter.Pending(ctx, "Nova %s not found", cluster.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	nova.AddReadyCheck(deps, cluster)

	if result, err := deps.Wait(ctx, reporter.Pending); err != nil || !result.IsZero() {
		return result, err
	}

	if err := computeset.Reconcile(ctx, r.Client, instance, log); err != nil {
		if err := reporter.Error(ctx, "Error reconciling Nova compute set: %v", err); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	cell := &openstackv1beta1.NovaCell{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(cluster.Name, instance.Spec.Cell),
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cell), cell); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := reporter.Pending(ctx, "NovaCell %s not found", cell.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	cinder := &openstackv1beta1.Cinder{
		// TODO make this configurable
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cinder",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cinder), cinder); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		cinder = nil
	}

	cm := computeset.ConfigMap(instance, cell, cluster, cinder)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	keystoneSecret := template.Combine(cluster.Name, "keystone")

	// TODO most of these are probably not needed
	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("KOLLA_SKIP_EXTEND_START", "true"),
		template.EnvVar("REQUESTS_CA_BUNDLE", "/etc/ssl/certs/ca-certificates.crt"),
		// TODO make ingress optional
		template.EnvVar("OS_VNC__NOVNCPROXY_BASE_URL", fmt.Sprintf("https://%s/vnc_auto.html", cell.Spec.NoVNCProxy.Ingress.Host)),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", cell.Spec.Broker.Secret, "connection"),
		template.SecretEnvVar("OS_DATABASE__CONNECTION", cell.Spec.Database.Secret, "connection"),
		template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__MEMCACHE_SECRET_KEY", "keystone-memcache", "secret-key"),
	}

	env = append(env, keystone.MiddlewareEnv("OS_KEYSTONE_AUTHTOKEN__", keystoneSecret)...)
	env = append(env, keystone.ClientEnv("OS_NEUTRON__", cluster.Spec.Neutron.Secret)...)
	env = append(env, keystone.ClientEnv("OS_PLACEMENT__", cluster.Spec.Placement.Secret)...)

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-nova", cm.Name, nil),
	}

	if err := r.reconcileLibvirtd(ctx, instance, cinder, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileCompute(ctx, instance, cell, cinder, env, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileComputeSSH(ctx, instance, configHash, volumes, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := reporter.Reconciled(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NovaComputeSetReconciler) reconcileLibvirtd(ctx context.Context, instance *openstackv1beta1.NovaComputeSet, cinder *openstackv1beta1.Cinder, log logr.Logger) error {
	cm := nova.LibvirtdConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("KOLLA_SKIP_EXTEND_START", "true"),
	}

	var (
		volumeMounts []corev1.VolumeMount
		volumes      []corev1.Volume
	)

	if cinder != nil {
		cephSecrets := rookceph.NewClientSecretAppender(&volumes, &volumeMounts)
		for _, backend := range cinder.Spec.Backends {
			if cephSpec := backend.Ceph; cephSpec != nil {
				cephSecrets.Append(cephSpec.Secret)

				// TODO support multiple ceph backends
				env = append(env, template.EnvVar("LIBVIRT_CEPH_CINDER_SECRET_UUID", "74a0b63e-041d-4040-9398-3704e4cf8260"))
				env = append(env, template.EnvVar("CEPH_CINDER_USER", cephSpec.ClientName))
				env = append(env, template.EnvVar("CEPH_CINDER_SECRET", cephSpec.Secret))
			}
		}
	}

	ds := nova.LibvirtdDaemonSet(instance, env, volumeMounts, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaComputeSetReconciler) reconcileCompute(ctx context.Context, instance *openstackv1beta1.NovaComputeSet, cell *openstackv1beta1.NovaCell, cinder *openstackv1beta1.Cinder, env []corev1.EnvVar, volumes []corev1.Volume, log logr.Logger) error {
	var volumeMounts []corev1.VolumeMount

	if cinder != nil {
		cephSecrets := rookceph.NewClientSecretAppender(&volumes, &volumeMounts)
		for _, backend := range cinder.Spec.Backends {
			if cephSpec := backend.Ceph; cephSpec != nil {
				cephSecrets.Append(cephSpec.Secret)

				// TODO support multiple ceph backends
				env = append(env, template.EnvVar("OS_LIBVIRT__RBD_SECRET_UUID", "74a0b63e-041d-4040-9398-3704e4cf8260"))
				env = append(env, template.EnvVar("OS_LIBVIRT__RBD_USER", cephSpec.ClientName))
			}
		}
	}

	ds := nova.ComputeDaemonSet(instance, cell.Spec.Broker, env, volumeMounts, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaComputeSetReconciler) reconcileComputeSSH(ctx context.Context, instance *openstackv1beta1.NovaComputeSet, configHash string, volumes []corev1.Volume, log logr.Logger) error {
	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
		template.EnvVar("KOLLA_SKIP_EXTEND_START", "true"),
	}

	ds := nova.ComputeSSHDaemonSet(instance, env, volumes)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaComputeSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	nodeRequestsFn := handler.MapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
		c := mgr.GetClient()

		var sets openstackv1beta1.NovaComputeSetList
		if err := c.List(context.Background(), &sets); err != nil {
			r.Log.Error(err, "Failed to list NovaComputeSet")
			return nil
		}

		var requests []reconcile.Request
		for _, instance := range sets.Items {
			// TODO limit to compute sets that matched the changed node
			name := client.ObjectKey{
				Namespace: instance.Namespace,
				Name:      instance.Name,
			}
			requests = append(requests, reconcile.Request{NamespacedName: name})
		}

		return requests
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaComputeSet{}).
		Watches(&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(nodeRequestsFn)).
		Owns(&openstackv1beta1.NovaComputeNode{}).
		Complete(r)
}
