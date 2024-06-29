package octavia

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Octavia, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.API
	driverAgentSpec := instance.Spec.DriverAgent

	runAsRootUser := int64(0)

	volumes = append(volumes,
		template.EmptyDirVolume("pod-var-run-octavia"))

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-octavia", "/etc/octavia/octavia.conf", "octavia.conf"),
		template.VolumeMount("pod-var-run-octavia", "/var/run/octavia"),
	}

	apiVolumeMounts := append(volumeMounts,
		template.SubPathVolumeMount("etc-octavia", "/etc/apache2/sites-available/000-default.conf", "httpd.conf"),
		template.SubPathVolumeMount("etc-octavia", "/var/lib/kolla/config_files/config.json", "kolla-octavia-api.json"))

	driverAgentVolumeMounts := append(volumeMounts,
		template.SubPathVolumeMount("etc-octavia", "/var/lib/kolla/config_files/config.json", "kolla-octavia-driver-agent.json"))

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(9876),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		FailureThreshold:    15,
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		InitContainers: []corev1.Container{
			{
				Name:  "init",
				Image: spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "init-driver-agent.sh"),
				},
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &runAsRootUser,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Containers: []corev1.Container{
			{
				Name:    "api",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9876},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &runAsRootUser,
				},
				VolumeMounts: apiVolumeMounts,
			},
			{
				Name:         "driver-agent",
				Image:        driverAgentSpec.Image,
				Command:      []string{"/usr/local/bin/kolla_start"},
				Env:          env,
				Resources:    driverAgentSpec.Resources,
				VolumeMounts: driverAgentVolumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "api")

	return deploy
}

func APIService(instance *openstackv1beta1.Octavia) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, "api")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 9876},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Octavia) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	name := template.Combine(instance.Name, "api")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.API.Ingress, labels)
}
