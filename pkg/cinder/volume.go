package cinder

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	VolumeComponentLabel = "volume"
)

func VolumeStatefulSet(instance *openstackv1beta1.Cinder, envVars []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, VolumeComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "etc-cinder",
			SubPath:   "cinder.conf",
			MountPath: "/etc/cinder/cinder.conf",
		},
	}

	if cephSpec := instance.Spec.Volume.Storage.RookCeph; cephSpec != nil {
		volumeMounts = append(volumeMounts, rookceph.ClientVolumeMounts("etc-ceph")...)
		volumes = append(volumes, template.SecretVolume("etc-ceph", cephSpec.Secret, nil))
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.Volume.Replicas,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:  "volume",
				Image: instance.Spec.Image,
				Command: []string{
					"cinder-volume",
					"--config-file=/etc/cinder/cinder.conf",
				},
				Env:          envVars,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "volume")

	return sts
}

func VolumeService(instance *openstackv1beta1.Cinder) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, VolumeComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "volume", "headless"),
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
