package neutron

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ServerComponentLabel = "server"
)

func ServerDeployment(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(9696),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron.conf", "neutron.conf"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Server.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "server",
				Image: instance.Spec.Image,
				Command: []string{
					"neutron-server",
					"--config-file=/etc/neutron/neutron.conf",
				},
				Env: env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9696},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts:   volumeMounts,
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "server")

	return deploy
}

func ServerService(instance *openstackv1beta1.Neutron) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)
	name := template.Combine(instance.Name, "server")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 9696},
	}

	return svc
}

func ServerIngress(instance *openstackv1beta1.Neutron) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	name := template.Combine(instance.Name, "server")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.Server.Ingress, labels)
}
