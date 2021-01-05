package template

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateJob(ctx context.Context, c client.Client, intended *batchv1.Job, log logr.Logger) error {
	found := &batchv1.Job{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Job", "Name", intended.Name)
		return c.Create(ctx, intended)
	}
	return nil
}
