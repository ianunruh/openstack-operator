package magnum

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ConductorComponentLabel = "conductor"
)

func ConductorStatefulSet(instance *openstackv1beta1.Magnum, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-magnum", "/etc/magnum/magnum.conf", "magnum.conf"),
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Conductor.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "conductor",
				Image: instance.Spec.Image,
				Command: []string{
					"magnum-conductor",
					"--config-file=/etc/magnum/magnum.conf",
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

	sts.Name = template.Combine(instance.Name, "conductor")

	return sts
}

func ConductorService(instance *openstackv1beta1.Magnum) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)
	name := template.Combine(instance.Name, "conductor", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
