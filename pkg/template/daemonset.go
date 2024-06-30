package template

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureDaemonSet(ctx context.Context, c client.Client, instance *appsv1.DaemonSet, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *appsv1.DaemonSet) {
		instance.Spec = intended.Spec
	})
}
