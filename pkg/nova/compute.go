package nova

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ComputeComponentLabel = "compute"
)

func Compute(instance *openstackv1beta1.NovaCell, name string, spec openstackv1beta1.NovaComputeSpec) *openstackv1beta1.NovaCompute {
	labels := template.AppLabels(instance.Name, AppLabel)

	spec.Cell = instance.Spec.Name

	return &openstackv1beta1.NovaCompute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, name),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
}

func EnsureCompute(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaCompute, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating NovaCompute", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating NovaCompute", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

func ComputeDaemonSet(instance *openstackv1beta1.NovaCompute, env []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume, containerImage string) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, ComputeComponentLabel)

	runAsRootUser := int64(0)
	privileged := true
	rootOnlyRootFilesystem := true

	initVolumeMounts := []corev1.VolumeMount{
		template.VolumeMount("pod-shared", "/tmp/pod-shared"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
	}

	extraVolumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.VolumeMount("pod-tmp", "/tmp"),
		template.VolumeMount("pod-shared", "/tmp/pod-shared"),
		template.VolumeMount("host-dev", "/dev"),
		template.ReadOnlyVolumeMount("host-etc-machine-id", "/etc/machine-id"),
		template.ReadOnlyVolumeMount("host-lib-modules", "/lib/modules"),
		template.VolumeMount("host-run", "/run"),
		template.ReadOnlyVolumeMount("host-sys-fs-cgroup", "/sys/fs/cgroup"),
		template.BidirectionalVolumeMount("host-var-lib-libvirt", "/var/lib/libvirt"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
	}

	extraVolumes := []corev1.Volume{
		template.EmptyDirVolume("pod-tmp"),
		template.EmptyDirVolume("pod-shared"),
		template.HostPathVolume("host-dev", "/dev"),
		template.HostPathVolume("host-etc-machine-id", "/etc/machine-id"),
		template.HostPathVolume("host-lib-modules", "/lib/modules"),
		template.HostPathVolume("host-run", "/run"),
		template.HostPathVolume("host-sys-fs-cgroup", "/sys/fs/cgroup"),
		template.HostPathVolume("host-var-lib-libvirt", "/var/lib/libvirt"),
		template.HostPathVolume("host-var-lib-nova", "/var/lib/nova"),
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
		},
		InitContainers: []corev1.Container{
			{
				Name:  "compute-init",
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "compute-init.sh"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("NOVA_USER_UID", strconv.Itoa(int(appUID))),
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  &runAsRootUser,
					Privileged: &privileged,
				},
				VolumeMounts: initVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "compute",
				Image: containerImage,
				Command: []string{
					"nova-compute",
					"--config-file=/etc/nova/nova.conf",
					"--config-file=/tmp/pod-shared/nova-hypervisor.conf",
				},
				Env: env,
				SecurityContext: &corev1.SecurityContext{
					Privileged:             &privileged,
					ReadOnlyRootFilesystem: &rootOnlyRootFilesystem,
				},
				VolumeMounts: append(volumeMounts, extraVolumeMounts...),
			},
		},
		Volumes: append(volumes, extraVolumes...),
	})

	ds.Name = template.Combine(instance.Name, "compute")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true

	return ds
}
