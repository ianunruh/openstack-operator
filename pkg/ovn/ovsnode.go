package ovn

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	OVSNodeComponentLabel = "ovs-node"
)

func OVSNodeDaemonSet(instance *openstackv1beta1.OVNControlPlane) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, OVSNodeComponentLabel)

	privileged := true

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.Node.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "openvswitch",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "start-node.sh"),
				},
				Env: []corev1.EnvVar{
					// TODO make configurable
					template.EnvVar("NIC", "eno1"),
					template.EnvVar("BRIDGE_MAPPINGS", "external:br-ex"),
					template.EnvVar("BRIDGE_PORTS", "br-ex:vlan3000"),
					template.EnvVar("GATEWAY", "true"),
					template.FieldEnvVar("HOSTNAME", "spec.nodeName"),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromConfigMap(template.Combine(instance.Name, "ovsdb")),
				},
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: []corev1.VolumeMount{
					template.ReadOnlyVolumeMount("host-lib-modules", "/lib/modules"),
					template.ReadOnlyVolumeMount("host-sys", "/sys"),
					template.VolumeMount("host-etc-openvswitch", "/etc/openvswitch"),
					template.VolumeMount("host-run-openvswitch", "/run/openvswitch"),
					template.VolumeMount("host-var-lib-openvswitch", "/var/lib/openvswitch"),
				},
			},
		},
		Volumes: []corev1.Volume{
			template.HostPathVolume("host-lib-modules", "/lib/modules"),
			template.HostPathVolume("host-sys", "/sys"),
			template.HostPathVolume("host-etc-openvswitch", "/etc/openvswitch"),
			template.HostPathVolume("host-run-openvswitch", "/run/openvswitch"),
			template.HostPathVolume("host-var-lib-openvswitch", "/var/lib/openvswitch"),
		},
	})

	ds.Name = template.Combine(instance.Name, "ovs-node")

	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true
	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

	return ds
}
