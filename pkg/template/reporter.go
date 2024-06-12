package template

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
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

	cw.Clear()

	return ctrl.Result{}
}

func (cw *ConditionWaiter) Clear() {
	cw.resources = nil
}

type conditionWaitResource struct {
	Instance   client.Object
	Conditions []metav1.Condition
}

func NewReporter(instance client.Object, conditions *[]metav1.Condition, k8sClient client.Client, recorder record.EventRecorder) *Reporter {
	return &Reporter{
		instance:   instance,
		conditions: conditions,

		client:   k8sClient,
		recorder: recorder,
	}
}

// Reporter provides a generic way to report the status of a resource.
type Reporter struct {
	instance   client.Object
	conditions *[]metav1.Condition

	client   client.Client
	recorder record.EventRecorder
}

func (r *Reporter) UpdateCondition(ctx context.Context, conditionType string, status metav1.ConditionStatus, reason, message string, args ...any) error {
	message = fmt.Sprintf(message, args...)

	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: r.instance.GetGeneration(),
	}
	if !meta.SetStatusCondition(r.conditions, condition) {
		return nil
	}

	r.recorder.Event(r.instance, corev1.EventTypeNormal, reason, message)

	if err := r.client.Status().Update(ctx, r.instance); err != nil {
		return fmt.Errorf("updating object status: %w", err)
	}

	return nil
}

func (r *Reporter) UpdateReadyCondition(ctx context.Context, status metav1.ConditionStatus, reason, message string, args ...any) error {
	return r.UpdateCondition(ctx, openstackv1beta1.ConditionReady, status, reason, message, args...)
}
