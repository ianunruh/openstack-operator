package senlin

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	EngineComponentLabel = "engine"
)

func EngineDeployment(instance *openstackv1beta1.Senlin, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, EngineComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-senlin", "/etc/senlin/senlin.conf", "senlin.conf"),
	}

	sts := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Engine.Replicas,
		NodeSelector: instance.Spec.Engine.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "engine",
				Image: instance.Spec.Image,
				Command: []string{
					"senlin-engine",
					"--config-file=/etc/senlin/senlin.conf",
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

	sts.Name = template.Combine(instance.Name, "engine")

	return sts
}