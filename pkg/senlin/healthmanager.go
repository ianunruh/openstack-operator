package senlin

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	HealthManagerComponentLabel = "health-manager"
)

func HealthManagerDeployment(instance *openstackv1beta1.Senlin, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, HealthManagerComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-senlin", "/etc/senlin/senlin.conf", "senlin.conf"),
		template.SubPathVolumeMount("etc-senlin", "/var/lib/kolla/config_files/config.json", "kolla-senlin-health-manager.json"),
	}

	sts := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.HealthManager.Replicas,
		NodeSelector: instance.Spec.HealthManager.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "health-manager",
				Image:        instance.Spec.Image,
				Command:      []string{"/usr/local/bin/kolla_start"},
				Env:          env,
				Resources:    instance.Spec.HealthManager.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "health-manager")

	return sts
}
