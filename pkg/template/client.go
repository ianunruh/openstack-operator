package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type CopyableObject[T client.Object] interface {
	client.Object
	DeepCopy() T
}

func Ensure[T CopyableObject[T]](ctx context.Context, c client.Client, instance T, log logr.Logger, update func(T)) error {
	hash, err := ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	gvk, err := apiutil.GVKForObject(instance, c.Scheme())
	if err != nil {
		return fmt.Errorf("mapping GVK for object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating resource",
			"Name", instance.GetName(),
			"Namespace", instance.GetNamespace(),
			"Kind", gvk.Kind)

		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		update(intended)

		SetAppliedHash(instance, hash)

		log.Info("Updating resource",
			"Name", instance.GetName(),
			"Namespace", instance.GetNamespace(),
			"Kind", gvk.Kind)

		return c.Update(ctx, instance)
	}

	return nil
}
