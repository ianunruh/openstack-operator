package rally

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ToolboxComponentLabel = "toolbox"
)

func ToolboxDeployment(instance *openstackv1beta1.Rally, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ToolboxComponentLabel)

	spec := instance.Spec.Toolbox

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-rally", "/home/rally/.rally/rally.conf", "rally.conf"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "toolbox",
				Image:        spec.Image,
				Command:      []string{"/bin/bash"},
				Env:          env,
				Resources:    spec.Resources,
				Stdin:        true,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "toolbox")

	return deploy
}
