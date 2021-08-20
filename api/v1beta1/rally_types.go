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

// RallySpec defines the desired state of Rally
type RallySpec struct {
	Image string `json:"image"`

	Database MariaDBDatabaseSpec `json:"database"`

	Data *VolumeSpec `json:"data"`
}

// RallyStatus defines the observed state of Rally
type RallyStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Rally is the Schema for the rallies API
type Rally struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RallySpec   `json:"spec,omitempty"`
	Status RallyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RallyList contains a list of Rally
type RallyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rally `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rally{}, &RallyList{})
}
