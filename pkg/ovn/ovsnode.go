package ovn

import (
	"strings"

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

	env := []corev1.EnvVar{
		template.FieldEnvVar("HOSTNAME", "spec.nodeName"),
	}

	cfg := instance.Spec.Node
	if len(cfg.BridgeMappings) > 0 {
		env = append(env,
			template.EnvVar("BRIDGE_MAPPINGS", strings.Join(cfg.BridgeMappings, ",")),
			template.EnvVar("BRIDGE_PORTS", strings.Join(cfg.BridgePorts, ",")),
			template.EnvVar("GATEWAY", "true"))
	}

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
				Env: env,
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
