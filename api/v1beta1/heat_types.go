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

// HeatSpec defines the desired state of Heat
type HeatSpec struct {
	Image string `json:"image"`

	// +optional
	API HeatAPISpec `json:"api"`

	// +optional
	CFN HeatAPISpec `json:"cfn"`

	// +optional
	Engine HeatEngineSpec `json:"engine"`

	Database MariaDBDatabaseSpec `json:"database"`

	Broker RabbitMQUserSpec `json:"broker"`
}

type HeatAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

type HeatEngineSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

// HeatStatus defines the observed state of Heat
type HeatStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Heat is the Schema for the heats API
type Heat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HeatSpec   `json:"spec,omitempty"`
	Status HeatStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HeatList contains a list of Heat
type HeatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Heat `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Heat{}, &HeatList{})
}
