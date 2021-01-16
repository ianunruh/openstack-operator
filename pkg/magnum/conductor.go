package magnum

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

func ConductorStatefulSet(instance *openstackv1beta1.Magnum, envVars []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)

	sts := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Conductor.Replicas,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:  "conductor",
				Image: instance.Spec.Image,
				Command: []string{
					"magnum-conductor",
					"--config-file=/etc/magnum/magnum.conf",
				},
				Env: envVars,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-magnum",
						SubPath:   "magnum.conf",
						MountPath: "/etc/magnum/magnum.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "conductor")

	return sts
}

func ConductorService(instance *openstackv1beta1.Magnum) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "conductor", "headless"),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}
