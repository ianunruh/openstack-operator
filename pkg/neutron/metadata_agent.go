package neutron

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	MetadataAgentComponentLabel = "metadata-agent"
)

func MetadataAgentDaemonSet(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, MetadataAgentComponentLabel)

	spec := instance.Spec.MetadataAgent

	privileged := true
	shareProcessNamespace := true

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron.conf", "neutron.conf"),
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron_ovn_metadata_agent.ini", "neutron_ovn_metadata_agent.ini"),
		template.BidirectionalVolumeMount("host-run-netns", "/run/netns"),
		template.SubPathVolumeMount("etc-neutron", "/var/lib/kolla/config_files/config.json", "kolla-neutron-metadata-agent.json"),
	}

	volumes = append(volumes,
		template.HostPathVolume("host-run-netns", "/run/netns"))

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:      "agent",
				Image:     spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Env:       env,
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "metadata-agent")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.ShareProcessNamespace = &shareProcessNamespace

	return ds
}
