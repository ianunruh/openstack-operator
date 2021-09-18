package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	SchedulerComponentLabel = "scheduler"
)

func SchedulerStatefulSet(instance *openstackv1beta1.Nova, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, SchedulerComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Scheduler.Replicas,
		NodeSelector: instance.Spec.Scheduler.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "scheduler",
				Image: instance.Spec.Image,
				Command: []string{
					"nova-scheduler",
					"--config-file=/etc/nova/nova.conf",
				},
				Env:          env,
				Resources:    instance.Spec.Scheduler.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "scheduler")

	return sts
}

func SchedulerService(instance *openstackv1beta1.Nova) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, SchedulerComponentLabel)
	name := template.Combine(instance.Name, "scheduler", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
