package manila

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ShareComponentLabel = "share"
)

func ShareStatefulSet(instance *openstackv1beta1.Manila, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ShareComponentLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-manila", "/etc/manila/manila.conf", "manila.conf"),
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
		Replicas:     instance.Spec.Share.Replicas,
		NodeSelector: instance.Spec.Share.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "share",
				Image: instance.Spec.Image,
				Command: []string{
					"manila-share",
					"--config-file=/etc/manila/manila.conf",
				},
				Env:          env,
				Resources:    instance.Spec.Share.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(instance.Name, "share")

	return sts
}

func ShareService(instance *openstackv1beta1.Manila) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ShareComponentLabel)
	name := template.Combine(instance.Name, "share", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
