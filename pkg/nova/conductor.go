package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ConductorComponentLabel = "conductor"
)

func ConductorStatefulSet(name, namespace string, spec openstackv1beta1.NovaConductorSpec, envVars []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *appsv1.StatefulSet {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)

	sts := template.GenericStatefulSet(template.Component{
		Namespace: namespace,
		Labels:    labels,
		Replicas:  spec.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "conductor",
				Image: containerImage,
				Command: []string{
					"nova-conductor",
					"--config-file=/etc/nova/nova.conf",
					"--debug",
				},
				Env: envVars,
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

	sts.Name = template.Combine(name, "conductor")

	return sts
}

func ConductorService(name, namespace string) *corev1.Service {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(name, "conductor", "headless"),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}
