package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HealthManagerComponentLabel = "health-manager"
)

func HealthManagerDaemonSet(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, HealthManagerComponentLabel)

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.HealthManager.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:  "manager",
				Image: instance.Spec.Image,
				Command: []string{
					"octavia-health-manager",
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

	ds.Name = template.Combine(instance.Name, "health-manager")

	return ds
}
