package barbican

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	WorkerComponentLabel = "worker"
)

func WorkerDeployment(instance *openstackv1beta1.Barbican, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, WorkerComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-barbican", "/etc/barbican/barbican.conf", "barbican.conf"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Worker.Replicas,
		NodeSelector: instance.Spec.Worker.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "worker",
				Image: instance.Spec.Image,
				Command: []string{
					"barbican-worker",
					"--config-file=/etc/barbican/barbican.conf",
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

	deploy.Name = template.Combine(instance.Name, "worker")

	return deploy
}
