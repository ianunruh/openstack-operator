package neutron

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	MetadataAgentComponentLabel = "metadata-agent"
)

func MetadataAgentDaemonSet(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, MetadataAgentComponentLabel)

	privileged := true

	extraVolumes := []corev1.Volume{
		template.HostPathVolume("host-run-netns", "/run/netns"),
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron.conf", "neutron.conf"),
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron_ovn_metadata_agent.ini", "neutron_ovn_metadata_agent.ini"),
		template.BidirectionalVolumeMount("host-run-netns", "/run/netns"),
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.MetadataAgent.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "agent",
				Image: instance.Spec.Image,
				Command: []string{
					"neutron-ovn-metadata-agent",
					"--config-file=/etc/neutron/neutron.conf",
					"--config-file=/etc/neutron/neutron_ovn_metadata_agent.ini",
				},
				Env:       env,
				Resources: instance.Spec.MetadataAgent.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: append(volumes, extraVolumes...),
	})

	ds.Name = template.Combine(instance.Name, "metadata-agent")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true

	return ds
}
