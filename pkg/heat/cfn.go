package heat

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	CFNComponentLabel = "cfn"
)

func CFNDeployment(instance *openstackv1beta1.Heat, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, CFNComponentLabel)

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(8000),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-heat", "/etc/heat/heat.conf", "heat.conf"),
		template.SubPathVolumeMount("etc-heat", "/var/lib/kolla/config_files/config.json", "kolla-heat-api-cfn.json"),
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.CFN.Replicas,
		NodeSelector: instance.Spec.CFN.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: []corev1.Container{
			{
				Name:    "cfn",
				Image:   instance.Spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 8000},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     instance.Spec.CFN.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "cfn")

	return deploy
}

func CFNService(instance *openstackv1beta1.Heat) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, CFNComponentLabel)
	name := template.Combine(instance.Name, "cfn")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 8000},
	}

	return svc
}

func CFNIngress(instance *openstackv1beta1.Heat) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, CFNComponentLabel)

	name := template.Combine(instance.Name, "cfn")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.CFN.Ingress, labels)
}
