package template

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateJob(ctx context.Context, c client.Client, instance *batchv1.Job, log logr.Logger) error {
	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Job", "Name", instance.Name)
		return c.Create(ctx, instance)
	}
	return nil
}
