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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MagnumSpec defines the desired state of Magnum
type MagnumSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	API MagnumAPISpec `json:"api,omitempty"`

	// +optional
	Conductor MagnumConductorSpec `json:"conductor,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type MagnumAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type MagnumConductorSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// MagnumStatus defines the observed state of Magnum
type MagnumStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=magnums
//+kubebuilder:subresource:status

// Magnum is the Schema for the magnums API
type Magnum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MagnumSpec   `json:"spec,omitempty"`
	Status MagnumStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MagnumList contains a list of Magnum
type MagnumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Magnum `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Magnum{}, &MagnumList{})
}
