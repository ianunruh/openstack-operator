package cinder

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/httpd"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Cinder, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.API

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromInt(8776),
				Scheme: pki.HTTPActionScheme(spec.TLS),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-cinder", "/etc/cinder/cinder.conf", "cinder.conf"),
		template.SubPathVolumeMount("etc-cinder", "/etc/apache2/sites-available/000-default.conf", "httpd.conf"),
		template.SubPathVolumeMount("etc-cinder", "/var/lib/kolla/config_files/config.json", "kolla-cinder-api.json"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)
	pki.AppendRabbitMQTLSClientVolumes(instance.Spec.Broker, &volumes, &volumeMounts)
	pki.AppendTLSServerVolumes(spec.TLS, "/etc/cinder/certs", 0400, &volumes, &volumeMounts)

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
				Name:      "api",
				Image:     spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Lifecycle: httpd.Lifecycle(),
				Env:       env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 8776},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     spec.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "api")

	return deploy
}

func APIService(instance *openstackv1beta1.Cinder) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, "api")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 8776},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Cinder) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	name := template.Combine(instance.Name, "api")
	spec := instance.Spec.API

	return template.GenericIngressWithTLS(name, instance.Namespace, spec.Ingress, spec.TLS, labels)
}

func APIInternalURL(instance *openstackv1beta1.Cinder, apiVersion string) string {
	scheme := "http"
	if instance.Spec.API.TLS.Secret != "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s-api.%s.svc:8776/%s/$(project_id)s", scheme, instance.Name, instance.Namespace, apiVersion)
}

func APIPublicURL(instance *openstackv1beta1.Cinder, apiVersion string) string {
	if instance.Spec.API.Ingress == nil {
		return APIInternalURL(instance, apiVersion)
	}
	return fmt.Sprintf("https://%s/%s/$(project_id)s", instance.Spec.API.Ingress.Host, apiVersion)
}
