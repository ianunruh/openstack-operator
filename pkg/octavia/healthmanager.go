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

func HealthManagerDeployment(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, HealthManagerComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
	}

	volumeMounts = append(volumeMounts, amphora.VolumeMounts(instance)...)
	volumes = append(volumes, amphora.Volumes(instance)...)

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.HealthManager.NodeSelector,
		// SecurityContext: &corev1.PodSecurityContext{
		// 	RunAsUser: &appUID,
		// 	FSGroup:   &appUID,
		// },
		InitContainers: []corev1.Container{
			amphora.InitContainer(instance.Spec.Image, volumeMounts),
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
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "health-manager")

	return deploy
}
