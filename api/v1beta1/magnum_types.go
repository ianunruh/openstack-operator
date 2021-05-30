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

// MagnumSpec defines the desired state of Magnum
type MagnumSpec struct {
	Image string `json:"image"`

	// +optional
	API MagnumAPISpec `json:"api"`

	// +optional
	Conductor MagnumConductorSpec `json:"conductor"`

	Database MariaDBDatabaseSpec `json:"database"`

	Broker RabbitMQUserSpec `json:"broker"`
}

type MagnumAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

type MagnumConductorSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

// MagnumStatus defines the observed state of Magnum
type MagnumStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Magnum is the Schema for the magnums API
type Magnum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MagnumSpec   `json:"spec,omitempty"`
	Status MagnumStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MagnumList contains a list of Magnum
type MagnumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Magnum `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Magnum{}, &MagnumList{})
}
