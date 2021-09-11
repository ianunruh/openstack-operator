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

// SenlinSpec defines the desired state of Senlin
type SenlinSpec struct {
	Image string `json:"image"`

	// +optional
	API SenlinAPISpec `json:"api,omitempty"`

	// +optional
	Conductor SenlinConductorSpec `json:"conductor,omitempty"`

	// +optional
	Engine SenlinEngineSpec `json:"engine,omitempty"`

	// +optional
	HealthManager SenlinHealthManagerSpec `json:"healthManager,omitempty"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type SenlinAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

type SenlinConductorSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type SenlinEngineSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type SenlinHealthManagerSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// SenlinStatus defines the observed state of Senlin
type SenlinStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Senlin is the Schema for the senlins API
type Senlin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SenlinSpec   `json:"spec,omitempty"`
	Status SenlinStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SenlinList contains a list of Senlin
type SenlinList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Senlin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Senlin{}, &SenlinList{})
}
