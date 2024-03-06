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
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/horizon/auth/login/",
				Port: intstr.FromInt(8080),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-horizon", "/etc/openstack-dashboard/local_settings", "local_settings.py"),
		template.SubPathVolumeMount("etc-horizon", "/var/lib/kolla/config_files/config.json", "kolla.json"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-horizon", instance.Name, nil),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Server.Replicas,
		NodeSelector: instance.Spec.Server.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: []corev1.Container{
			{
				Name:      "server",
				Image:     instance.Spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Lifecycle: httpd.Lifecycle(),
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.EnvVar("KOLLA_CONFIG_STRATEGY", "COPY_ALWAYS"),
					template.EnvVar("ENABLE_HEAT", "yes"),
					template.EnvVar("ENABLE_MANILA", "yes"),
					template.EnvVar("ENABLE_OCTAVIA", "yes"),
				},
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 80},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     instance.Spec.Server.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		Volumes: volumes,
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
