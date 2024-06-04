package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
)

type JobSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}
