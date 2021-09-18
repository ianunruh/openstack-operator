package ovn

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	NorthdComponentLabel = "northd"
)

func NorthdDeployment(instance *openstackv1beta1.OVNControlPlane) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, NorthdComponentLabel)

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Northd.Replicas,
		NodeSelector: instance.Spec.Northd.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: []corev1.Container{
			{
				Name:  "northd",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "start-northd.sh"),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromConfigMap(template.Combine(instance.Name, "ovsdb")),
				},
				Resources: instance.Spec.Northd.Resources,
			},
		},
	})

	deploy.Name = template.Combine(instance.Name, "northd")

	return deploy
}
