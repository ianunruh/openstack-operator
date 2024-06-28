package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureDeployment(ctx context.Context, c client.Client, instance *appsv1.Deployment, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(context.TODO(), client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating Deployment", "Name", intended.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		SetAppliedHash(instance, hash)

		log.Info("Updating Deployment", "Name", intended.Name)
		return c.Update(ctx, instance)
	}

	return nil
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
