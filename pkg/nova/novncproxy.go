package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	NoVNCProxyComponentLabel = "novncproxy"
)

func NoVNCProxyDeployment(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, NoVNCProxyComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.NoVNCProxy.Replicas,
		NodeSelector: instance.Spec.NoVNCProxy.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "novncproxy",
				Image: containerImage,
				Command: []string{
					"nova-novncproxy",
					"--config-file=/etc/nova/nova.conf",
				},
				Env: env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 6080},
				},
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "novncproxy")

	return deploy
}

func NoVNCProxyService(instance *openstackv1beta1.NovaCell) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, NoVNCProxyComponentLabel)
	name := template.Combine(instance.Name, "novncproxy")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 6080},
	}

	return svc
}

func NoVNCProxyIngress(instance *openstackv1beta1.NovaCell) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, "novncproxy")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.NoVNCProxy.Ingress, labels)
}
