package mariadb

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	readyMessage = "MariaDB is running"
)

func NewReporter(recorder record.EventRecorder) *Reporter {
	return &Reporter{
		recorder: recorder,
	}
}

type Reporter struct {
	recorder record.EventRecorder
}

func (r *Reporter) Pending(instance *openstackv1beta1.MariaDB, err error, eventReason, message string) {
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}

	// suppress duplicate pending events
	oldCondition := ReadyCondition(instance)
	if oldCondition == nil || oldCondition.Reason != openstackv1beta1.ReasonPending {
		r.recorder.Event(instance, corev1.EventTypeNormal, eventReason, message)
	}

	SetCondition(instance, openstackv1beta1.ConditionReady,
		metav1.ConditionFalse, openstackv1beta1.ReasonPending, message)
}

func (r *Reporter) Running(instance *openstackv1beta1.MariaDB) {
	r.recorder.Event(instance, corev1.EventTypeNormal, "MariaDBRunning", readyMessage)
	SetCondition(instance, openstackv1beta1.ConditionReady,
		metav1.ConditionTrue, openstackv1beta1.ReasonRunning, readyMessage)
}

func ReadyCondition(instance *openstackv1beta1.MariaDB) *metav1.Condition {
	return meta.FindStatusCondition(instance.Status.Conditions, openstackv1beta1.ConditionReady)
}

func SetCondition(instance *openstackv1beta1.MariaDB, conditionType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: instance.Generation,
	})
}

func AddReadyCheck(cw *template.ConditionWaiter, instance *openstackv1beta1.MariaDB) {
	cw.AddReadyCheck(instance, instance.Status.Conditions)
}
