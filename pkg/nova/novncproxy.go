package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	NoVNCProxyComponentLabel = "novncproxy"
)

func NoVNCProxyDeployment(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, NoVNCProxyComponentLabel)

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.NoVNCProxy.Replicas,
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
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-nova",
						SubPath:   "nova.conf",
						MountPath: "/etc/nova/nova.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "novncproxy")

	return deploy
}

func NoVNCProxyService(instance *openstackv1beta1.NovaCell) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, NoVNCProxyComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "novncproxy"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 6080},
			},
		},
	}

	return svc
}

func NoVNCProxyIngress(instance *openstackv1beta1.NovaCell) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.NoVNCProxy.Ingress

	prefixPathType := netv1.PathTypePrefix

	svcName := template.Combine(instance.Name, "novncproxy")

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        template.Combine(instance.Name, "novncproxy"),
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: spec.Annotations,
		},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{
				{
					SecretName: template.Combine(instance.Name, "novncproxy-ingress-tls"),
					Hosts:      []string{spec.Host},
				},
			},
			Rules: []netv1.IngressRule{
				{
					Host: spec.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									PathType: &prefixPathType,
									Path:     "/",
									Backend:  template.IngressServiceBackend(svcName, "http"),
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}
