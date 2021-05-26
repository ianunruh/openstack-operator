package keystone

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
		Replicas:  instance.Spec.API.Replicas,
		Containers: []corev1.Container{
			{
				Name:      "api",
				Image:     instance.Spec.Image,
				Command:   httpd.Command(),
				Lifecycle: httpd.Lifecycle(),
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
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
	name := template.Combine(instance.Name, "api")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 5000},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Keystone) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	name := template.Combine(instance.Name, "api")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.API.Ingress, labels)
}
