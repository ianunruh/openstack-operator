package task

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	PendingReason   = "RallyTaskPending"
	CompletedReason = "RallyTaskCompleted"
)

func NewReporter(instance *openstackv1beta1.RallyTask, k8sClient client.Client, recorder record.EventRecorder) *Reporter {
	return &Reporter{
		reporter: template.NewReporter(instance, &instance.Status.Conditions, k8sClient, recorder),
	}
}

type Reporter struct {
	reporter *template.Reporter
}

func (r *Reporter) Pending(ctx context.Context, message string, args ...any) error {
	return r.reporter.UpdateCompletedCondition(ctx, metav1.ConditionFalse, PendingReason, message, args...)
}

func (r *Reporter) Completed(ctx context.Context) error {
	return r.reporter.UpdateCompletedCondition(ctx, metav1.ConditionTrue, CompletedReason, "RallyTask is complete")
}
