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

func OVSNodeDaemonSet(instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, OVSNodeComponentLabel)

	spec := instance.Spec.Node

	env = append(env,
		template.FieldEnvVar("HOSTNAME", "spec.nodeName"),
		template.EnvVar("OVERLAY_CIDRS", strings.Join(spec.OverlayCIDRs, ",")))

	if len(spec.BridgeMappings) > 0 {
		env = append(env,
			template.EnvVar("BRIDGE_MAPPINGS", strings.Join(spec.BridgeMappings, ",")),
			template.EnvVar("BRIDGE_PORTS", strings.Join(spec.BridgePorts, ",")),
			template.EnvVar("GATEWAY", "true"))
	}

	volumeMounts := []corev1.VolumeMount{
		template.ReadOnlyVolumeMount("host-lib-modules", "/lib/modules"),
		template.ReadOnlyVolumeMount("host-sys", "/sys"),
		template.VolumeMount("host-etc-openvswitch", "/etc/openvswitch"),
		template.VolumeMount("host-run-openvswitch", "/run/openvswitch"),
		template.VolumeMount("host-var-lib-openvswitch", "/var/lib/openvswitch"),
	}

	dbVolumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-ovn", "/var/lib/kolla/config_files/config.json", "kolla-openvswitch-ovsdb.json"),
	}

	switchVolumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-ovn", "/var/lib/kolla/config_files/config.json", "kolla-openvswitch-vswitchd.json"),
	}

	volumes = append(volumes,
		template.HostPathVolume("host-lib-modules", "/lib/modules"),
		template.HostPathVolume("host-sys", "/sys"),
		template.HostPathVolume("host-etc-openvswitch", "/etc/openvswitch"),
		template.HostPathVolume("host-run-openvswitch", "/run/openvswitch"),
		template.HostPathVolume("host-var-lib-openvswitch", "/var/lib/openvswitch"))

	privileged := true

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.Node.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "ovsdb",
				Image:        spec.DB.Image,
				Command:      []string{"/usr/local/bin/kolla_start"},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: append(volumeMounts, dbVolumeMounts...),
			},
			{
				Name:      "vswitchd",
				Image:     spec.Switch.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Env:       env,
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: append(volumeMounts, switchVolumeMounts...),
			},
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "ovs-node")

	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true
	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

	return ds
}
