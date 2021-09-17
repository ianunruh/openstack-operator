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
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

// NovaComputeReconciler reconciles a NovaCompute object
type NovaComputeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacomputes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacomputes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openstack.k8s.ianunruh.com,resources=novacomputes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NovaCompute object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *NovaComputeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &openstackv1beta1.NovaCompute{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster := &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		return ctrl.Result{}, err
	}

	cell := &openstackv1beta1.NovaCell{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(cluster.Name, instance.Spec.Cell),
			Namespace: instance.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cell), cell); err != nil {
		return ctrl.Result{}, err
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

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cm), cm); err != nil {
		return ctrl.Result{}, err
	}
	configHash := template.AppliedHash(cm)

	keystoneSecret := template.Combine(cluster.Name, "keystone")

	// TODO most of these are probably not needed
	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", cluster.Spec.Broker.Secret, "connection"),
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

	if err := r.reconcileCompute(ctx, instance, cell, cinder, env, volumes, cluster.Spec.Image, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileComputeSSH(ctx, instance, cluster.Spec.Image, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NovaComputeReconciler) reconcileLibvirtd(ctx context.Context, instance *openstackv1beta1.NovaCompute, cinder *openstackv1beta1.Cinder, log logr.Logger) error {
	cm := nova.LibvirtdConfigMap(instance)
	controllerutil.SetControllerReference(instance, cm, r.Scheme)
	if err := template.EnsureConfigMap(ctx, r.Client, cm, log); err != nil {
		return err
	}
	configHash := template.AppliedHash(cm)

	env := []corev1.EnvVar{
		template.EnvVar("CONFIG_HASH", configHash),
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

func (r *NovaComputeReconciler) reconcileCompute(ctx context.Context, instance *openstackv1beta1.NovaCompute, cell *openstackv1beta1.NovaCell, cinder *openstackv1beta1.Cinder, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string, log logr.Logger) error {
	extraEnvVars := []corev1.EnvVar{
		template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", cell.Spec.Broker.Secret, "connection"),
		// TODO make ingress optional
		template.EnvVar("OS_VNC__NOVNCPROXY_BASE_URL", fmt.Sprintf("https://%s/vnc_auto.html", cell.Spec.NoVNCProxy.Ingress.Host)),
	}

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

	env = append(env, extraEnvVars...)

	ds := nova.ComputeDaemonSet(instance, env, volumeMounts, volumes, containerImage)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

func (r *NovaComputeReconciler) reconcileComputeSSH(ctx context.Context, instance *openstackv1beta1.NovaCompute, containerImage string, log logr.Logger) error {
	ds := nova.ComputeSSHDaemonSet(instance, containerImage)
	controllerutil.SetControllerReference(instance, ds, r.Scheme)
	if err := template.EnsureDaemonSet(ctx, r.Client, ds, log); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NovaComputeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1beta1.NovaCompute{}).
		Complete(r)
}
