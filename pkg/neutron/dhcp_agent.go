package neutron

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	DHCPAgentComponentLabel = "dhcp-agent"
)

func DHCPAgentDaemonSet(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, DHCPAgentComponentLabel)

	shareProcessNamespace := true

	readOnlyRootFilesystem := true
	privileged := true

	mountPropagation := corev1.MountPropagationBidirectional

	extraVolumes := []corev1.Volume{
		template.EmptyDirVolume("pod-tmp"),
		template.EmptyDirVolume("pod-shared"),
		template.EmptyDirVolume("pod-var-lib-neutron"),
		template.HostPathVolume("host-run-netns", "/run/netns"),
		// expected that metadata agent is on same host
		template.HostPathVolume("host-var-lib-neutron-metadata-proxy", "/var/lib/neutron/metadata-proxy"),
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.DHCPAgent.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "agent",
				Image: instance.Spec.Image,
				Command: []string{
					"neutron-dhcp-agent",
					"--config-file=/etc/neutron/neutron.conf",
					"--config-file=/etc/neutron/plugins/ml2/ml2_conf.ini",
					"--config-file=/etc/neutron/dhcp_agent.ini",
				},
				Env: env,
				VolumeMounts: []corev1.VolumeMount{
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
						Name:             "host-run-netns",
						MountPath:        "/run/netns",
						MountPropagation: &mountPropagation,
					},
					{
						Name:      "host-var-lib-neutron-metadata-proxy",
						MountPath: "/var/lib/neutron/metadata-proxy",
					},
				},
				SecurityContext: &corev1.SecurityContext{
					Privileged:             &privileged,
					ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
				},
			},
		},
		Volumes: append(volumes, extraVolumes...),
	})

	ds.Name = template.Combine(instance.Name, "dhcp-agent")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.ShareProcessNamespace = &shareProcessNamespace

	return ds
}
