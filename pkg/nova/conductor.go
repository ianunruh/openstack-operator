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

func ConductorStatefulSet(name, namespace string, spec openstackv1beta1.NovaConductorSpec, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-conductor.json"),
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:    "conductor",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				LivenessProbe: &corev1.Probe{
					ProbeHandler:        healthProbeHandler("conductor", true),
					InitialDelaySeconds: 120,
					PeriodSeconds:       90,
					TimeoutSeconds:      70,
				},
				StartupProbe: &corev1.Probe{
					ProbeHandler:        healthProbeHandler("conductor", false),
					InitialDelaySeconds: 80,
					PeriodSeconds:       90,
					TimeoutSeconds:      70,
				},
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
