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

func GenericConfigMap(name, namespace string, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{},
	}
}

func EnsureConfigMap(ctx context.Context, c client.Client, intended *corev1.ConfigMap, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	SetAppliedHash(intended, hash)

	found := &corev1.ConfigMap{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating ConfigMap", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		found.Data = intended.Data
		SetAppliedHash(found, hash)

		log.Info("Updating ConfigMap", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
