package cinder

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	VolumeComponentLabel = "volume"
)

func VolumeStatefulSet(instance *openstackv1beta1.Cinder, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, VolumeComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-cinder", "/etc/cinder/cinder.conf", "cinder.conf"),
	}

	cephSecrets := rookceph.NewClientSecretAppender(&volumes, &volumeMounts)
	for _, backend := range instance.Spec.Backends {
		if cephSpec := backend.Ceph; cephSpec != nil {
			cephSecrets.Append(cephSpec.Secret)
		}
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.Volume.Replicas,
		NodeSelector: instance.Spec.Volume.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "volume",
				Image: instance.Spec.Image,
				Command: []string{
					"cinder-volume",
					"--config-file=/etc/cinder/cinder.conf",
				},
				Env:          env,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "volume")

	return sts
}

func VolumeService(instance *openstackv1beta1.Cinder) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, VolumeComponentLabel)
	name := template.Combine(instance.Name, "volume", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
