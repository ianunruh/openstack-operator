package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	MetadataComponentLabel = "metadata"
)

func MetadataDeployment(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, MetadataComponentLabel)

	env = append(env, template.EnvVar("OS_DEFAULT__ENABLED_APIS", "metadata"))

	// probe := &corev1.Probe{
	// 	Handler: corev1.Handler{
	// 		HTTPGet: &corev1.HTTPGetAction{
	// 			Path: "/",
	// 			Port: intstr.FromInt(8775),
	// 		},
	// 	},
	// 	InitialDelaySeconds: 10,
	// 	PeriodSeconds:       10,
	// 	TimeoutSeconds:      5,
	// }

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Metadata.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "metadata",
				Image: containerImage,
				Command: []string{
					"nova-api",
					"--config-file=/etc/nova/nova.conf",
				},
				Env: env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 8775},
				},
				// LivenessProbe:  probe,
				// ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-nova",
						SubPath:   "nova.conf",
						MountPath: "/etc/nova/nova.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "metadata")

	return deploy
}

func MetadataService(instance *openstackv1beta1.NovaCell) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, MetadataComponentLabel)
	name := template.Combine(instance.Name, "metadata")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 8775},
	}

	return svc
}
