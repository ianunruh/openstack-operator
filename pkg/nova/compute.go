package nova

import (
	"slices"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ComputeComponentLabel = "compute"
)

func ComputeDaemonSet(instance *openstackv1beta1.NovaComputeSet, brokerSpec openstackv1beta1.RabbitMQUserSpec, env []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, ComputeComponentLabel)

	runAsRootUser := int64(0)
	privileged := true

	probe := &corev1.Probe{
		ProbeHandler:        amqpHealthProbeHandler("nova-compute", brokerSpec),
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	initVolumeMounts := []corev1.VolumeMount{
		template.VolumeMount("pod-shared", "/tmp/pod-shared"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
	}

	volumeMounts = slices.Concat(volumeMounts, []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-compute.json"),
		template.VolumeMount("pod-tmp", "/tmp"),
		template.VolumeMount("pod-shared", "/tmp/pod-shared"),
		template.VolumeMount("host-dev", "/dev"),
		template.ReadOnlyVolumeMount("host-etc-machine-id", "/etc/machine-id"),
		template.ReadOnlyVolumeMount("host-lib-modules", "/lib/modules"),
		template.VolumeMount("host-run", "/run"),
		template.ReadOnlyVolumeMount("host-sys-fs-cgroup", "/sys/fs/cgroup"),
		template.BidirectionalVolumeMount("host-var-lib-libvirt", "/var/lib/libvirt"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
	})

	volumes = slices.Concat(volumes, []corev1.Volume{
		template.EmptyDirVolume("pod-tmp"),
		template.EmptyDirVolume("pod-shared"),
		template.HostPathVolume("host-dev", "/dev"),
		template.HostPathVolume("host-etc-machine-id", "/etc/machine-id"),
		template.HostPathVolume("host-lib-modules", "/lib/modules"),
		template.HostPathVolume("host-run", "/run"),
		template.HostPathVolume("host-sys-fs-cgroup", "/sys/fs/cgroup"),
		template.HostPathVolume("host-var-lib-libvirt", "/var/lib/libvirt"),
		template.HostPathVolume("host-var-lib-nova", "/var/lib/nova"),
	})

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)
	pki.AppendRabbitMQTLSClientVolumes(brokerSpec, &volumes, &volumeMounts)

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
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "compute-init.sh"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("NOVA_USER_UID", strconv.Itoa(int(appUID))),
				},
				Resources: instance.Spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  &runAsRootUser,
					Privileged: &privileged,
				},
				VolumeMounts: initVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:          "compute",
				Image:         instance.Spec.Image,
				Command:       []string{"/usr/local/bin/kolla_start"},
				Env:           env,
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     instance.Spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "compute")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true

	return ds
}
