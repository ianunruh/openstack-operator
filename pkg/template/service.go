package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HeadlessServiceName(name string) string {
	return Combine(name, "headless")
}

func GenericService(name, namespace string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
		},
	}
}

func EnsureService(ctx context.Context, c client.Client, instance *corev1.Service, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(intended, hash)

		log.Info("Creating Service", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(instance, hash) {
		// copy immutable fields
		intended.Spec.ClusterIP = instance.Spec.ClusterIP

		instance.Spec = intended.Spec
		SetAppliedHash(instance, hash)

		log.Info("Updating Service", "Name", intended.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
