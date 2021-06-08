package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ComputeComponentLabel = "compute"
)

func ComputeDaemonSet(instance *openstackv1beta1.Nova, envVars []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, ComputeComponentLabel)

	runAsRootUser := int64(0)
	runAsNovaUser := int64(64060)
	privileged := true
	rootOnlyRootFilesystem := true

	mountPropagation := corev1.MountPropagationBidirectional

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

	extraVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "etc-nova",
			MountPath: "/etc/nova/nova.conf",
			SubPath:   "nova.conf",
		},
		{
			Name:      "pod-tmp",
			MountPath: "/tmp",
		},
		{
			Name:      "pod-shared",
			MountPath: "/tmp/pod-shared",
		},
		{
			Name:      "host-dev",
			MountPath: "/dev",
		},
		{
			Name:      "host-etc-machine-id",
			MountPath: "/etc/machine-id",
			ReadOnly:  true,
		},
		{
			Name:      "host-lib-modules",
			MountPath: "/lib/modules",
			ReadOnly:  true,
		},
		{
			Name:      "host-run",
			MountPath: "/run",
		},
		{
			Name:      "host-sys-fs-cgroup",
			MountPath: "/sys/fs/cgroup",
			ReadOnly:  true,
		},
		{
			Name:             "host-var-lib-libvirt",
			MountPath:        "/var/lib/libvirt",
			MountPropagation: &mountPropagation,
		},
		{
			Name:             "host-var-lib-nova",
			MountPath:        "/var/lib/nova",
			MountPropagation: &mountPropagation,
		},
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.Compute.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &runAsNovaUser,
		},
		InitContainers: []corev1.Container{
			{
				Name:  "compute-init",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "compute-init.sh"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("NOVA_USER_UID", "64060"),
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  &runAsRootUser,
					Privileged: &privileged,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "pod-shared",
						MountPath: "/tmp/pod-shared",
					},
					{
						Name:             "host-var-lib-nova",
						MountPath:        "/var/lib/nova",
						MountPropagation: &mountPropagation,
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "compute",
				Image: instance.Spec.Image,
				Command: []string{
					"nova-compute",
					"--config-file=/etc/nova/nova.conf",
					"--config-file=/tmp/pod-shared/nova-hypervisor.conf",
				},
				Env: envVars,
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
