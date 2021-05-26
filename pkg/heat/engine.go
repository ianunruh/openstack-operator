package heat

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	EngineComponentLabel = "engine"
)

func EngineStatefulSet(instance *openstackv1beta1.Heat, envVars []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, EngineComponentLabel)

	sts := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Engine.Replicas,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:  "engine",
				Image: instance.Spec.Image,
				Command: []string{
					"heat-engine",
					"--config-file=/etc/heat/heat.conf",
				},
				Env: envVars,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-heat",
						SubPath:   "heat.conf",
						MountPath: "/etc/heat/heat.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "engine")

	return sts
}

func EngineService(instance *openstackv1beta1.Heat) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, EngineComponentLabel)
	name := template.Combine(instance.Name, "engine", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
