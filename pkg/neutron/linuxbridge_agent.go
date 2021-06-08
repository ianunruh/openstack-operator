package neutron

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	LinuxBridgeAgentComponentLabel = "linuxbridge-agent"
)

func LinuxBridgeAgentDaemonSet(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, LinuxBridgeAgentComponentLabel)

	shareProcessNamespace := true

	runAsRootUser := int64(0)
	readOnlyRootFilesystem := true
	privileged := true

	extraVolumes := []corev1.Volume{
		template.EmptyDirVolume("pod-tmp"),
		template.EmptyDirVolume("pod-shared"),
		template.EmptyDirVolume("pod-var-lib-neutron"),
		template.HostPathVolume("host-rootfs", "/"),
		template.HostPathVolume("host-run", "/run"),
		template.HostPathVolume("host-var-lib-ebtables", "/var/lib/ebtables"),
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "etc-neutron",
			MountPath: "/etc/neutron/neutron.conf",
			SubPath:   "neutron.conf",
		},
		{
			Name:      "etc-neutron",
			MountPath: "/etc/neutron/plugins/ml2/ml2_conf.ini",
			SubPath:   "ml2_conf.ini",
		},
		{
			Name:      "etc-neutron",
			MountPath: "/etc/neutron/plugins/ml2/linuxbridge_agent.ini",
			SubPath:   "linuxbridge_agent.ini",
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
			Name:      "pod-var-lib-neutron",
			MountPath: "/var/lib/neutron",
		},
		{
			Name:      "host-run",
			MountPath: "/run",
		},
		{
			Name:      "host-var-lib-ebtables",
			MountPath: "/var/lib/ebtables",
		},
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.LinuxBridgeAgent.NodeSelector,
		InitContainers: []corev1.Container{
			{
				Name:  "init-modules",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "linuxbridge-agent-init-modules.sh"),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "host-rootfs",
						MountPath: "/mnt/host-rootfs",
						ReadOnly:  true,
					},
				},
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"SYS_MODULE",
							"SYS_CHROOT",
						},
					},
					ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
					RunAsUser:              &runAsRootUser,
				},
			},
			{
				Name:  "init-agent",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "linuxbridge-agent-init.sh"),
				},
				VolumeMounts: volumeMounts,
				SecurityContext: &corev1.SecurityContext{
					Privileged:             &privileged,
					ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
					RunAsUser:              &runAsRootUser,
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "agent",
				Image: instance.Spec.Image,
				Command: []string{
					"neutron-linuxbridge-agent",
					"--config-file=/etc/neutron/neutron.conf",
					"--config-file=/etc/neutron/plugins/ml2/ml2_conf.ini",
					"--config-file=/etc/neutron/plugins/ml2/linuxbridge_agent.ini",
					"--config-file=/tmp/pod-shared/ml2-local-ip.ini",
				},
				Env:          env,
				VolumeMounts: volumeMounts,
				SecurityContext: &corev1.SecurityContext{
					Privileged:             &privileged,
					ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
				},
			},
		},
		Volumes: append(volumes, extraVolumes...),
	})

	ds.Name = template.Combine(instance.Name, "linuxbridge-agent")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.ShareProcessNamespace = &shareProcessNamespace

	return ds
}
