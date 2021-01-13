package horizon

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				Port: intstr.FromInt(80),
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

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "server"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80},
			},
		},
	}

	return svc
}

func ServerIngress(instance *openstackv1beta1.Horizon) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	spec := instance.Spec.Server.Ingress

	prefixPathType := netv1.PathTypePrefix

	svcName := template.Combine(instance.Name, "server")

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        template.Combine(instance.Name, "server"),
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: template.MergeStringMaps(annotations, spec.Annotations),
		},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{
				{
					SecretName: template.Combine(instance.Name, "server-ingress-tls"),
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
