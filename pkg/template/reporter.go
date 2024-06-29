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

type ReportFunc func(ctx context.Context, message string, args ...any) error

func NewConditionWaiter(log logr.Logger) *ConditionWaiter {
	return &ConditionWaiter{
		log: log,
	}
}

type ConditionWaiter struct {
	log logr.Logger

	resources []conditionWaitResource
}

func (cw *ConditionWaiter) AddCheck(instance client.Object, conditionType string, ready bool) *ConditionWaiter {
	cw.resources = append(cw.resources, conditionWaitResource{
		Instance:      instance,
		ConditionType: conditionType,
		Ready:         ready,
	})
	return cw
}

func (cw *ConditionWaiter) AddReadyCheck(instance client.Object, conditions []metav1.Condition) *ConditionWaiter {
	return cw.AddCheck(instance,
		openstackv1beta1.ConditionReady,
		meta.IsStatusConditionTrue(conditions, openstackv1beta1.ConditionReady))
}

func (cw *ConditionWaiter) Wait(ctx context.Context, report ReportFunc) (ctrl.Result, error) {
	for _, res := range cw.resources {
		if res.Ready {
			continue
		}

		if err := report(
			ctx,
			"Waiting on %s %s condition %s",
			res.Instance.GetObjectKind().GroupVersionKind().Kind,
			res.Instance.GetName(),
			res.ConditionType,
		); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	cw.resources = nil

	return ctrl.Result{}, nil
}

type conditionWaitResource struct {
	Instance      client.Object
	ConditionType string
	Ready         bool
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

func (r *Reporter) UpdateCompletedCondition(ctx context.Context, status metav1.ConditionStatus, reason, message string, args ...any) error {
	return r.UpdateCondition(ctx, openstackv1beta1.ConditionCompleted, status, reason, message, args...)
}
