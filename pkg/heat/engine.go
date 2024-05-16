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

func EngineStatefulSet(instance *openstackv1beta1.Heat, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, EngineComponentLabel)

	spec := instance.Spec.Engine

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-heat", "/etc/heat/heat.conf", "heat.conf"),
		template.SubPathVolumeMount("etc-heat", "/var/lib/kolla/config_files/config.json", "kolla-heat-engine.json"),
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "engine",
				Image:        spec.Image,
				Command:      []string{"/usr/local/bin/kolla_start"},
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
