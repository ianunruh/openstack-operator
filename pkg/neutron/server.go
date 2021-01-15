package neutron

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
	ServerComponentLabel = "server"
)

func ServerDeployment(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(9696),
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
				Name:  "server",
				Image: instance.Spec.Image,
				Command: []string{
					"neutron-server",
					"--config-file=/etc/neutron/neutron.conf",
					"--config-file=/etc/neutron/plugins/ml2/ml2_conf.ini",
				},
				Env: env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9696},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-neutron",
						MountPath: "/etc/neutron/neutron.conf",
						SubPath:   "neutron.conf",
					},
					{
						Name:      "etc-neutron",
						MountPath: "/etc/neutron/plugins/ml2/ml2_conf.ini",
						SubPath:   "ml2_conf.ini",
					},
				},
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "server")

	return deploy
}

func ServerService(instance *openstackv1beta1.Neutron) *corev1.Service {
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
				{Name: "http", Port: 9696},
			},
		},
	}

	return svc
}

func ServerIngress(instance *openstackv1beta1.Neutron) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ServerComponentLabel)

	name := template.Combine(instance.Name, "server")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.Server.Ingress, labels)
}
