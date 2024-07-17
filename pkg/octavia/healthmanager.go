package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/octavia/amphora"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HealthManagerComponentLabel = "health-manager"
)

func HealthManagerDaemonSet(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, HealthManagerComponentLabel)

	spec := instance.Spec.HealthManager

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
		template.SubPathVolumeMount("etc-octavia", "/var/lib/kolla/config_files/config.json", "kolla-octavia-health-manager.json"),
	}

	defaultMode := int32(0400)

	// openvswitch volumes
	initVolumeMounts := []corev1.VolumeMount{
		template.VolumeMount("host-var-run-openvswitch", "/var/run/openvswitch"),
		template.VolumeMount("pod-shared", "/tmp/pod-shared"),
		template.SubPathVolumeMount("keystone", "/etc/openstack/clouds.yaml", "clouds.yaml"),
	}
	volumeMounts = append(volumeMounts, initVolumeMounts...)
	volumes = append(volumes,
		template.EmptyDirVolume("pod-shared"),
		template.HostPathVolume("host-var-run-openvswitch", "/var/run/openvswitch"),
		template.SecretVolume("keystone", "octavia-keystone", &defaultMode))

	// pki volumes
	volumeMounts = append(volumeMounts, amphora.VolumeMounts(instance)...)
	volumes = append(volumes, amphora.Volumes(instance)...)

	// XXX wire this into initVolumeMounts
	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	hmNetworkID := instance.Status.Amphora.NetworkIDs[0]

	initEnv := append(env,
		template.FieldEnvVar("HOSTNAME", "spec.nodeName"),
		template.EnvVar("HM_IFACE", "o-hm0"),
		template.EnvVar("HM_NETWORK_ID", hmNetworkID))

	privileged := true
	runAsRootUser := int64(0)

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: spec.NodeSelector,
		InitContainers: []corev1.Container{
			amphora.InitContainer(spec.Image, spec.Resources, volumeMounts),
			{
				Name:  "init-port",
				Image: spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-health-manager-port.sh"),
				},
				Env:          initEnv,
				Resources:    spec.Resources,
				VolumeMounts: initVolumeMounts,
			},
			{
				Name:  "init-ovs",
				Image: spec.InitOVSImage,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-health-manager-ovs.sh"),
				},
				Env:       initEnv,
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
					RunAsUser:  &runAsRootUser,
				},
				VolumeMounts: initVolumeMounts,
			},
			{
				Name:  "init-dhcp",
				Image: spec.InitDHCPImage,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-health-manager-dhcp.sh"),
				},
				Env:       initEnv,
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
					RunAsUser:  &runAsRootUser,
				},
				VolumeMounts: initVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:      "manager",
				Image:     spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Env:       env,
				Resources: spec.Resources,
				Ports: []corev1.ContainerPort{
					{Name: "heartbeat", ContainerPort: 5555, Protocol: corev1.ProtocolUDP},
				},
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "health-manager")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true

	return ds
}
