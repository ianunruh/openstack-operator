package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureDaemonSet(ctx context.Context, c client.Client, instance *appsv1.DaemonSet, log logr.Logger) error {
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

		log.Info("Creating DaemonSet", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		SetAppliedHash(instance, hash)

		log.Info("Updating DaemonSet", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
