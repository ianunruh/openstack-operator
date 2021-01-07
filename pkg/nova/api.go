package nova

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

func APIDeployment(instance *openstackv1beta1.Nova, configHash string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	keystoneSecret := template.Combine(instance.Name, "keystone")
	placementKeystoneSecret := "placement-keystone"

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(8774),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.API.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "api",
				Image: instance.Spec.Image,
				Command: []string{
					"nova-api",
					"--config-file=/etc/nova/nova.conf",
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.EnvVar("OS_DEFAULT__ENABLED_APIS", "osapi_compute"),
					template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
					template.SecretEnvVar("OS_API_DATABASE__CONNECTION", instance.Spec.APIDatabase.Secret, "connection"),
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
					template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneSecret, "OS_PASSWORD"),
					template.SecretEnvVar("OS_PLACEMENT__PASSWORD", placementKeystoneSecret, "OS_PASSWORD"),
				},
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 8774},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-nova",
						SubPath:   "nova.conf",
						MountPath: "/etc/nova/nova.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-nova", instance.Name, nil),
		},
	})

	deploy.Name = template.Combine(instance.Name, "api")

	return deploy
}

func APIService(instance *openstackv1beta1.Nova) *corev1.Service {
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
				{Name: "http", Port: 8774},
			},
		},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Nova) *netv1.Ingress {
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
