package placement

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

func APIDeployment(instance *openstackv1beta1.Placement, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.API

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromInt(8778),
				Scheme: pki.HTTPActionScheme(spec.TLS),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-placement", "/etc/apache2/sites-available/000-default.conf", "httpd.conf"),
		template.SubPathVolumeMount("etc-placement", "/etc/placement/placement.conf", "placement.conf"),
		template.SubPathVolumeMount("etc-placement", "/var/lib/kolla/config_files/config.json", "kolla.json"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)
	pki.AppendTLSServerVolumes(spec.TLS, "/etc/placement/certs", 0400, &volumes, &volumeMounts)

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
					{Name: "http", ContainerPort: 8778},
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

func APIService(instance *openstackv1beta1.Placement) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, "api")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 8778},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Placement) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	name := template.Combine(instance.Name, "api")
	spec := instance.Spec.API

	return template.GenericIngressWithTLS(name, instance.Namespace, spec.Ingress, spec.TLS, labels)
}

func APIInternalURL(instance *openstackv1beta1.Placement) string {
	scheme := "http"
	if instance.Spec.API.TLS.Secret != "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s-api.%s.svc:8778", scheme, instance.Name, instance.Namespace)
}

func APIPublicURL(instance *openstackv1beta1.Placement) string {
	if instance.Spec.API.Ingress == nil {
		return APIInternalURL(instance)
	}
	return fmt.Sprintf("https://%s", instance.Spec.API.Ingress.Host)
}
