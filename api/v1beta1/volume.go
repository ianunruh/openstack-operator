package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
)

type VolumeSpec struct {
	// +optional
	Capacity string `json:"capacity,omitempty"`

	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
}
