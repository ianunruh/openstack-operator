package user

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func NewReporter(instance *openstackv1beta1.KeystoneUser, k8sClient client.Client, recorder record.EventRecorder) *Reporter {
	return &Reporter{
		reporter: template.NewReporter(instance, &instance.Status.Conditions, k8sClient, recorder),
	}
}

type Reporter struct {
	reporter *template.Reporter
}

func (r *Reporter) Error(ctx context.Context, message string, args ...any) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionFalse, openstackv1beta1.ReasonReconcileError, message, args...)
}

func (r *Reporter) DeleteError(ctx context.Context, message string, args ...any) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionFalse, openstackv1beta1.ReasonDeleteError, message, args...)
}

func (r *Reporter) Pending(ctx context.Context, message string, args ...any) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionFalse, openstackv1beta1.ReasonPending, message, args...)
}

func (r *Reporter) Reconciled(ctx context.Context) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionTrue, openstackv1beta1.ReasonReconciled, "KeystoneUser is reconciled")
}

func AddReadyCheck(cw *template.ConditionWaiter, instance *openstackv1beta1.KeystoneUser) {
	cw.AddReadyCheck(instance, instance.Status.Conditions)
}
