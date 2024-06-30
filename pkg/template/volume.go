package template

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func PersistentVolumeClaim(name string, labels map[string]string, spec openstackv1beta1.VolumeSpec) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      spec.AccessModes,
			StorageClassName: spec.StorageClass,
			Resources:        StorageResources(spec.Capacity),
		},
	}
}

func EnsurePersistentVolumeClaim(ctx context.Context, c client.Client, instance *corev1.PersistentVolumeClaim, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *corev1.PersistentVolumeClaim) {
		instance.Spec = intended.Spec
	})
}
