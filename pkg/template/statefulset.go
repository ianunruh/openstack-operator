package template

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureStatefulSet(ctx context.Context, c client.Client, instance *appsv1.StatefulSet, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *appsv1.StatefulSet) {
		instance.Spec = intended.Spec
	})
}

func AddStatefulSetReadyCheck(cw *ConditionWaiter, instance *appsv1.StatefulSet) {
	cw.AddCheck(instance, "Available", instance.Status.AvailableReplicas > 0)
}
