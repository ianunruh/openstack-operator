package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/octavia/amphora"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HealthManagerComponentLabel = "health-manager"
)

func HealthManagerDaemonSet(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, HealthManagerComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
	}

	// openvswitch volumes
	initVolumeMounts := []corev1.VolumeMount{
		template.VolumeMount("host-var-run-openvswitch", "/var/run/openvswitch"),
	}
	volumeMounts = append(volumeMounts, initVolumeMounts...)
	volumes = append(volumes,
		template.HostPathVolume("host-var-run-openvswitch", "/var/run/openvswitch"))

	// pki volumes
	volumeMounts = append(volumeMounts, amphora.VolumeMounts(instance)...)
	volumes = append(volumes, amphora.Volumes(instance)...)

	// TODO support multiple replicas of health manager
	port := instance.Status.Amphora.HealthPorts[0]

	env = append(env,
		template.EnvVar("OS_HEALTH_MANAGER__BIND_IP", port.IPAddress))

	privileged := true
	runAsRootUser := int64(0)

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.HealthManager.NodeSelector,
		InitContainers: []corev1.Container{
			amphora.InitContainer(instance.Spec.Image, volumeMounts),
			{
				Name:  "init-port",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-health-manager-port.sh"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("HM_PORT_ID", port.ID),
					template.EnvVar("HM_PORT_MAC", port.MACAddress),
				},
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
					RunAsUser:  &runAsRootUser,
				},
				VolumeMounts: initVolumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "manager",
				Image: instance.Spec.Image,
				Command: []string{
					"octavia-health-manager",
					"--config-file=/etc/octavia/octavia.conf",
				},
				Env:          env,
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
