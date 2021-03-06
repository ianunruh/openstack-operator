package barbican

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
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Barbican, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(9311),
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
				Env:       env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9311},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-barbican",
						SubPath:   "barbican.conf",
						MountPath: "/etc/barbican/barbican.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "api")

	return deploy
}

func APIService(instance *openstackv1beta1.Barbican) *corev1.Service {
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
				{Name: "http", Port: 9311},
			},
		},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Barbican) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	name := template.Combine(instance.Name, "api")

	return template.GenericIngress(name, instance.Namespace, instance.Spec.API.Ingress, labels)
}
