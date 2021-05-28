package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func PersistentVolumeClaim(name string, labels map[string]string, spec openstackv1beta1.VolumeSpec) corev1.PersistentVolumeClaim {
	accessModes := spec.AccessModes
	if accessModes == nil {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      accessModes,
			StorageClassName: spec.StorageClass,
			Resources:        StorageResources(spec.Capacity),
		},
	}
}

func EnsurePersistentVolumeClaim(ctx context.Context, c client.Client, instance *corev1.PersistentVolumeClaim, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating PersistentVolumeClaim", "Name", intended.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		SetAppliedHash(instance, hash)

		log.Info("Updating PersistentVolumeClaim", "Name", intended.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
