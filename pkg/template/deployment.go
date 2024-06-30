package template

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureDeployment(ctx context.Context, c client.Client, instance *appsv1.Deployment, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *appsv1.Deployment) {
		instance.Spec = intended.Spec
	})
}

func AddDeploymentReadyCheck(cw *ConditionWaiter, instance *appsv1.Deployment) {
	cw.AddCheck(instance,
		string(appsv1.DeploymentAvailable),
		isDeploymentConditionStatusTrue(instance.Status.Conditions, appsv1.DeploymentAvailable))
}

func isDeploymentConditionStatusTrue(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
