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

func healthProbeHandler(queueName string, liveness bool) corev1.ProbeHandler {
	cmd := []string{
		"python3",
		"/usr/local/bin/nova-health-probe",
		"--config-file",
		"/etc/nova/nova.conf",
		"--service-queue-name",
		queueName,
	}

	if liveness {
		cmd = append(cmd, "--liveness-probe")
	}

	return corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: cmd,
		},
	}
}
