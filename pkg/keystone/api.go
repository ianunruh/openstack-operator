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

func APIDeployment(instance *openstackv1beta1.Keystone, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/v3/",
				Port: intstr.FromInt(5000),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-keystone", "/etc/apache2/sites-available/000-default.conf", "httpd.conf"),
		template.SubPathVolumeMount("etc-keystone", "/etc/keystone/keystone.conf", "keystone.conf"),
		template.SubPathVolumeMount("etc-keystone", "/var/lib/kolla/config_files/config.json", "kolla.json"),
		template.VolumeMount("pod-credential-keys", "/etc/keystone/credential-keys"),
		template.VolumeMount("pod-fernet-keys", "/etc/keystone/fernet-keys"),
	}

	initVolumeMounts := []corev1.VolumeMount{
		template.VolumeMount("credential-keys", "/var/run/secrets/credential-keys"),
		template.VolumeMount("fernet-keys", "/var/run/secrets/fernet-keys"),
	}

	extraVolumes := []corev1.Volume{
		template.EmptyDirVolume("pod-credential-keys"),
		template.EmptyDirVolume("pod-fernet-keys"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.API.Replicas,
		NodeSelector: instance.Spec.API.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		InitContainers: []corev1.Container{
			{
				Name:  "init-keys",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-keys.sh"),
				},
				Resources:    instance.Spec.API.Resources,
				VolumeMounts: append(volumeMounts, initVolumeMounts...),
			},
		},
		Containers: []corev1.Container{
			{
				Name:      "api",
				Image:     instance.Spec.Image,
				Command:   []string{"/usr/local/bin/kolla_start"},
				Lifecycle: httpd.Lifecycle(),
				Env:       env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 5000},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     instance.Spec.API.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		Volumes: append(volumes, extraVolumes...),
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
