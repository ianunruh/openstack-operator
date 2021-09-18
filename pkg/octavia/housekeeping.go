package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/octavia/amphora"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HousekeepingComponentLabel = "housekeeping"
)

func HousekeepingDeployment(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, HousekeepingComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
	}

	volumeMounts = append(volumeMounts, amphora.VolumeMounts(instance)...)
	volumes = append(volumes, amphora.Volumes(instance)...)

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Housekeeping.Replicas,
		NodeSelector: instance.Spec.Housekeeping.NodeSelector,
		InitContainers: []corev1.Container{
			amphora.InitContainer(instance.Spec.Image, instance.Spec.Housekeeping.Resources, volumeMounts),
		},
		Containers: []corev1.Container{
			{
				Name:  "housekeeping",
				Image: instance.Spec.Image,
				Command: []string{
					"octavia-housekeeping",
					"--config-file=/etc/octavia/octavia.conf",
				},
				Env:       env,
				Resources: instance.Spec.Housekeeping.Resources,
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

	deploy.Name = template.Combine(instance.Name, "housekeeping")

	deploy.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType

	deploy.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	deploy.Spec.Template.Spec.HostNetwork = true

	return deploy
}
