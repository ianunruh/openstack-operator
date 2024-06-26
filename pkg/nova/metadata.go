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

func MetadataDeployment(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, MetadataComponentLabel)

	spec := instance.Spec.Metadata

	env = append(env, template.EnvVar("OS_DEFAULT__ENABLED_APIS", "metadata"))

	// probe := &corev1.Probe{
	// 	ProbeHandler: corev1.ProbeHandler{
	// 		HTTPGet: &corev1.HTTPGetAction{
	// 			Path: "/",
	// 			Port: intstr.FromInt(8775),
	// 		},
	// 	},
	// 	InitialDelaySeconds: 5,
	// 	PeriodSeconds:       10,
	// 	TimeoutSeconds:      5,
	// }

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-api.json"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:    "metadata",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 8775},
				},
				// LivenessProbe: probe,
				// StartupProbe:  probe,
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
