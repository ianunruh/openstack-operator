/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GlanceSpec defines the desired state of Glance
type GlanceSpec struct {
	Image string `json:"image"`

	// +optional
	API GlanceAPISpec `json:"api"`

	Database MariaDBDatabaseSpec `json:"database"`

	// +optional
	Backends []GlanceBackendSpec `json:"backends"`
}

type GlanceBackendSpec struct {
	Name string `json:"name"`

	// +optional
	Default bool `json:"default"`

	// +optional
	Ceph *CephSpec `json:"ceph"`

	// +optional
	PVC *VolumeSpec `json:"pvc"`
}

type GlanceAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

// GlanceStatus defines the observed state of Glance
type GlanceStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Glance is the Schema for the glances API
type Glance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlanceSpec   `json:"spec,omitempty"`
	Status GlanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlanceList contains a list of Glance
type GlanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Glance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Glance{}, &GlanceList{})
}
