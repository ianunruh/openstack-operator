package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HousekeepingComponentLabel = "housekeeping"
)

func HousekeepingDeployment(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, HousekeepingComponentLabel)

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Housekeeping.Replicas,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:  "housekeeping",
				Image: instance.Spec.Image,
				Command: []string{
					"octavia-housekeeping",
					"--config-file=/etc/octavia/octavia.conf",
				},
				Env: env,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-octavia",
						SubPath:   "octavia.conf",
						MountPath: "/etc/octavia/octavia.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "housekeeping")

	return deploy
}
