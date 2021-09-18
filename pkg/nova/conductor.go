package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ConductorComponentLabel = "conductor"
)

func ConductorStatefulSet(name, namespace string, spec openstackv1beta1.NovaConductorSpec, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *appsv1.StatefulSet {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "conductor",
				Image: containerImage,
				Command: []string{
					"nova-conductor",
					"--config-file=/etc/nova/nova.conf",
				},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(name, "conductor")

	return sts
}

func ConductorService(name, namespace string) *corev1.Service {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)
	name = template.Combine(name, "conductor")

	svc := template.GenericService(name, namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
