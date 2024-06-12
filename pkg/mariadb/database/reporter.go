package database

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	PendingReason    = "MariaDBDatabasePending"
	ReconciledReason = "MariaDBDatabaseReconciled"
)

func NewReporter(instance *openstackv1beta1.MariaDBDatabase, k8sClient client.Client, recorder record.EventRecorder) *Reporter {
	return &Reporter{
		reporter: template.NewReporter(instance, &instance.Status.Conditions, k8sClient, recorder),
	}
}

type Reporter struct {
	reporter *template.Reporter
}

func (r *Reporter) Pending(ctx context.Context, message string, args ...any) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionFalse, PendingReason, message, args...)
}

func (r *Reporter) Reconciled(ctx context.Context) error {
	return r.reporter.UpdateReadyCondition(ctx, metav1.ConditionTrue, ReconciledReason, "MariaDBDatabase is reconciled")
}

func AddReadyCheck(cw *template.ConditionWaiter, instance *openstackv1beta1.MariaDBDatabase) {
	cw.AddReadyCheck(instance, instance.Status.Conditions)
}
