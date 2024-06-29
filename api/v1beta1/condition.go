package v1beta1

const (
	ConditionReady     = "Ready"
	ConditionCompleted = "Completed"

	ReasonCompleted      = "Completed"
	ReasonDeleteError    = "DeleteError"
	ReasonPending        = "Pending"
	ReasonReconciled     = "Reconciled"
	ReasonReconcileError = "ReconcileError"
	ReasonRunning        = "Running"
)
