package ovn

import (
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ControllerComponentLabel = "controller"
)

func ControllerDaemonSet(instance *openstackv1beta1.OVNControlPlane, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, ControllerComponentLabel)

	spec := instance.Spec.Controller
	nodeSpec := instance.Spec.Node

	privileged := true

	env = append(env,
		template.FieldEnvVar("HOSTNAME", "spec.nodeName"))

	setupEnv := append(env, setupNodeEnv(nodeSpec)...)

	envFrom := []corev1.EnvFromSource{
		template.EnvFromConfigMap(template.Combine(instance.Name, "ovsdb")),
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-ovn", "/var/lib/kolla/config_files/config.json", "kolla-controller.json"),
		template.VolumeMount("host-etc-openvswitch", "/etc/openvswitch"),
		template.VolumeMount("host-run-openvswitch", "/run/openvswitch"),
		template.VolumeMount("host-var-lib-openvswitch", "/var/lib/openvswitch"),
	}

	initVolumeMounts := append(volumeMounts,
		template.SubPathVolumeMount("etc-ovn", "/scripts/get-encap-ip.py", "get-encap-ip.py"),
		template.SubPathVolumeMount("etc-ovn", "/scripts/setup-node.sh", "setup-node.sh"))

	volumes = append(volumes,
		template.HostPathVolume("host-etc-openvswitch", "/etc/openvswitch"),
		template.HostPathVolume("host-run-openvswitch", "/run/openvswitch"),
		template.HostPathVolume("host-var-lib-openvswitch", "/var/lib/openvswitch"))

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: nodeSpec.NodeSelector,
		InitContainers: []corev1.Container{
			{
				Name:         "setup",
				Image:        spec.Image,
				Command:      []string{"bash", "/scripts/setup-node.sh"},
				Env:          setupEnv,
				EnvFrom:      envFrom,
				Resources:    spec.Resources,
				VolumeMounts: initVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:      "controller",
				Image:     spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Env:       env,
				EnvFrom:   envFrom,
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "controller")

	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true
	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

	return ds
}

func setupNodeEnv(spec openstackv1beta1.OVNNodeSpec) []corev1.EnvVar {
	gateway := len(spec.BridgeMappings) > 0

	env := []corev1.EnvVar{
		template.EnvVar("OVERLAY_CIDRS", strings.Join(spec.OverlayCIDRs, ",")),
		template.EnvVar("GATEWAY", strconv.FormatBool(gateway)),
	}

	if gateway {
		env = append(env,
			template.EnvVar("BRIDGE_MAPPINGS", strings.Join(spec.BridgeMappings, ",")),
			template.EnvVar("BRIDGE_PORTS", strings.Join(spec.BridgePorts, ",")))
	}

	return env
}
