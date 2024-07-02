package neutron

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ServerComponentLabel = "server"
)

func ServerDeployment(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	spec := instance.Spec.Server

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromInt(9696),
				Scheme: pki.HTTPActionScheme(spec.TLS),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-neutron", "/etc/neutron/neutron.conf", "neutron.conf"),
		template.SubPathVolumeMount("etc-neutron", "/var/lib/kolla/config_files/config.json", "kolla-neutron-server.json"),
	}

	pki.AppendTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)
	pki.AppendTLSServerVolumes(spec.TLS, "/etc/neutron/certs", &volumes, &volumeMounts)

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
				Name:    "server",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9696},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     spec.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
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
	spec := instance.Spec.Server

	return template.GenericIngressWithTLS(name, instance.Namespace, spec.Ingress, spec.TLS, labels)
}

func ServerInternalURL(instance *openstackv1beta1.Neutron) string {
	scheme := "http"
	if instance.Spec.Server.TLS.Secret != "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s-server.%s.svc:9696", scheme, instance.Name, instance.Namespace)
}

func ServerPublicURL(instance *openstackv1beta1.Neutron) string {
	if instance.Spec.Server.Ingress == nil {
		return ServerInternalURL(instance)
	}
	return fmt.Sprintf("https://%s", instance.Spec.Server.Ingress.Host)
}
