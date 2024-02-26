package nova

import (
	corev1 "k8s.io/api/core/v1"
)

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
