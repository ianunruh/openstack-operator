package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureDeployment(ctx context.Context, c client.Client, intended *appsv1.Deployment, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &appsv1.Deployment{}
	if err := c.Get(context.TODO(), client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(intended, hash)

		log.Info("Creating Deployment", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		SetAppliedHash(found, hash)

		log.Info("Updating Deployment", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
