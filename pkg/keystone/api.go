package keystone

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Keystone, configHash string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/v3/",
				Port: intstr.FromInt(5000),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "api",
				Image: instance.Spec.Image,
				Command: []string{
					"apachectl",
					"-DFOREGROUND",
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"apachectl",
								"-k",
								"graceful-stop",
							},
						},
					},
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
				},
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 5000},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-keystone",
						SubPath:   "keystone.conf",
						MountPath: "/etc/keystone/keystone.conf",
					},
					{
						Name:      "credential-keys",
						MountPath: "/etc/keystone/credential-keys",
					},
					{
						Name:      "fernet-keys",
						MountPath: "/etc/keystone/fernet-keys",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-keystone", instance.Name, nil),
			template.SecretVolume("credential-keys", template.Combine(instance.Name, "credential-keys"), nil),
			template.SecretVolume("fernet-keys", template.Combine(instance.Name, "fernet-keys"), nil),
		},
	})

	deploy.Name = template.Combine(instance.Name, "api")

	return deploy
}

func APIService(instance *openstackv1beta1.Keystone) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "api"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 5000},
			},
		},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Keystone) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.API.Ingress

	prefixPathType := netv1.PathTypePrefix

	svcName := template.Combine(instance.Name, "api")

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        template.Combine(instance.Name, "api"),
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: spec.Annotations,
		},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{
				{
					SecretName: template.Combine(instance.Name, "api-ingress-tls"),
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
