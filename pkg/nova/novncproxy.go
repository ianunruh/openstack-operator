package nova

import (
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	NoVNCProxyComponentLabel = "novncproxy"
)

func NoVNCProxyDeployment(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, NoVNCProxyComponentLabel)

	spec := instance.Spec.NoVNCProxy

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-novncproxy.json"),
	}

	pki.AppendTLSServerVolumes(spec.TLS, "/etc/nova/certs", 0444, &volumes, &volumeMounts)
	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	if spec.TLS.Secret != "" {
		env = slices.Concat(env, []corev1.EnvVar{
			template.EnvVar("OS_DEFAULT__SSL_ONLY", "true"),
			template.EnvVar("OS_DEFAULT__CERT", "/etc/nova/certs/tls.crt"),
			template.EnvVar("OS_DEFAULT__KEY", "/etc/nova/certs/tls.key"),
		})
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:    "novncproxy",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 6080},
				},
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

	spec := instance.Spec.NoVNCProxy

	return template.GenericIngressWithTLS(name, instance.Namespace, spec.Ingress, spec.TLS, labels)
}
