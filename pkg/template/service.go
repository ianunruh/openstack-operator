package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HeadlessServiceName(name string) string {
	return Combine(name, "headless")
}

func EnsureService(ctx context.Context, c client.Client, intended *corev1.Service, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &corev1.Service{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(intended, hash)

		log.Info("Creating Service", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		// copy immutable fields
		intended.Spec.ClusterIP = found.Spec.ClusterIP

		found.Spec = intended.Spec
		SetAppliedHash(found, hash)

		log.Info("Updating Service", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
