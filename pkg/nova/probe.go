package nova

import (
	corev1 "k8s.io/api/core/v1"
)

const amqpPort = "5672"

func amqpHealthProbeHandler(processName string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"healthcheck_port", processName, amqpPort},
		},
	}
}
