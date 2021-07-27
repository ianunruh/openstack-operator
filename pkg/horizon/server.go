package horizon

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/httpd"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ServerComponentLabel = "server"
)

func ServerDeployment(instance *openstackv1beta1.Horizon, configHash string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/horizon/auth/login/",
				Port: intstr.FromInt(8080),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Server.Replicas,
		Containers: []corev1.Container{
			{
				Name:      "server",
				Image:     instance.Spec.Image,
				Command:   httpd.Command(),
				Lifecycle: httpd.Lifecycle(),
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
				},
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 80},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-horizon",
						SubPath:   "local_settings.py",
						MountPath: "/etc/openstack-dashboard/local_settings.py",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-horizon", instance.Name, nil),
		},
	})

	deploy.Name = template.Combine(instance.Name, "server")

	return deploy
}

func ServerService(instance *openstackv1beta1.Horizon) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)
	name := template.Combine(instance.Name, "server")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 8080},
	}

	return svc
}

func ServerIngress(instance *openstackv1beta1.Horizon) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)
	name := template.Combine(instance.Name, ServerComponentLabel)

	ingress := template.GenericIngress(name, instance.Namespace, instance.Spec.Server.Ingress, labels)
	ingress.Annotations = template.MergeStringMaps(ingress.Annotations, map[string]string{
		"nginx.ingress.kubernetes.io/app-root":        "/horizon",
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	})

	return ingress
}
