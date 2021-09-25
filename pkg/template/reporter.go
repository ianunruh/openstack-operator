package template

import (
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func NewConditionWaiter(log logr.Logger) *ConditionWaiter {
	return &ConditionWaiter{
		log: log,
	}
}

type ConditionWaiter struct {
	log logr.Logger

	resources []conditionWaitResource
}

func (cw *ConditionWaiter) AddReadyCheck(instance client.Object, conditions []metav1.Condition) *ConditionWaiter {
	cw.resources = append(cw.resources, conditionWaitResource{
		Instance:   instance,
		Conditions: conditions,
	})
	return cw
}

func (cw *ConditionWaiter) Wait() ctrl.Result {
	for _, res := range cw.resources {
		if meta.IsStatusConditionTrue(res.Conditions, openstackv1beta1.ConditionReady) {
			continue
		}

		cw.log.Info("Waiting for dependency to be ready",
			"kind", res.Instance.GetObjectKind().GroupVersionKind().Kind,
			"name", res.Instance.GetName(),
			"namespace", res.Instance.GetNamespace())

		return ctrl.Result{RequeueAfter: 10 * time.Second}
	}

	return ctrl.Result{}
}

type conditionWaitResource struct {
	Instance   client.Object
	Conditions []metav1.Condition
}
