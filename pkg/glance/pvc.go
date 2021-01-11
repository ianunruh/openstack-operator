package glance

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PersistentVolumeClaim(instance *openstackv1beta1.Glance) *corev1.PersistentVolumeClaim {
	labels := template.AppLabels(instance.Name, AppLabel)
	spec := instance.Spec.Storage.Volume

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      spec.AccessModes,
			Resources:        template.StorageResources(spec.Capacity),
			StorageClassName: spec.StorageClass,
		},
	}
}
