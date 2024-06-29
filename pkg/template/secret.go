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

func GenericSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{},
	}
}

func CreateSecret(ctx context.Context, c client.Client, instance *corev1.Secret, log logr.Logger) error {
	err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Secret", "Name", instance.Name)
		return c.Create(ctx, instance)
	}

	return nil
}

func EnsureSecret(ctx context.Context, c client.Client, instance *corev1.Secret, log logr.Logger) error {
	hash, err := ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating Secret", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Data = intended.Data
		instance.StringData = intended.StringData
		SetAppliedHash(instance, hash)

		log.Info("Updating Secret", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
