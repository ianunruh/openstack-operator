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

func ServerDeployment(instance *openstackv1beta1.Horizon, env []corev1.EnvVar) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	spec := instance.Spec.Server

	probeHandler := corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/auth/login/",
			Port: intstr.FromInt(8080),
		},
	}

	livenessProbe := &corev1.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	startupProbe := &corev1.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 10,
		FailureThreshold:    30,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-horizon", "/etc/apache2/sites-available/000-default.conf", "httpd.conf"),
		template.SubPathVolumeMount("etc-horizon", "/etc/openstack-dashboard/local_settings", "local_settings.py"),
		template.SubPathVolumeMount("etc-horizon", "/var/lib/kolla/config_files/config.json", "kolla.json"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-horizon", instance.Name, nil),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: []corev1.Container{
			{
				Name:      "server",
				Image:     spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Lifecycle: httpd.Lifecycle(),
				Env:       env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 80},
				},
				LivenessProbe: livenessProbe,
				StartupProbe:  startupProbe,
				Resources:     spec.Resources,
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
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	})

	return ingress
}
