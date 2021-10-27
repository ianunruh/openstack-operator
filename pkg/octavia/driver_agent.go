package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	DriverAgentComponentLabel = "driver-agent"
)

func DriverAgentDeployment(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, DriverAgentComponentLabel)

	runAsRootUser := int64(0)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
		template.VolumeMount("host-var-run-octavia", "/var/run/octavia"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.DriverAgent.Replicas,
		NodeSelector: instance.Spec.DriverAgent.NodeSelector,
		InitContainers: []corev1.Container{
			{
				Name:  "init",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-driver-agent.sh"),
				},
				Resources: instance.Spec.DriverAgent.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &runAsRootUser,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "agent",
				Image: instance.Spec.Image,
				Command: []string{
					"octavia-driver-agent",
					"--config-file=/etc/octavia/octavia.conf",
				},
				Env:          env,
				Resources:    instance.Spec.DriverAgent.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "driver-agent")

	return deploy
}
