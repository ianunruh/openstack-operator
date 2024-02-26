package httpd

import (
	corev1 "k8s.io/api/core/v1"
)

func Command() []string {
	return []string{
		"apachectl",
		"-DFOREGROUND",
	}
}

func Lifecycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"apachectl",
					"-k",
					"graceful-stop",
				},
			},
		},
	}
}
