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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CinderSpec defines the desired state of Cinder
type CinderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image string `json:"image"`

	API CinderAPISpec `json:"api"`

	// +optional
	Scheduler CinderSchedulerSpec `json:"scheduler"`

	Database MariaDBDatabaseSpec `json:"database"`

	Broker RabbitMQUserSpec `json:"broker"`
}

type CinderAPISpec struct {
	// +optional
	Replicas int32        `json:"replicas"`
	Ingress  *IngressSpec `json:"ingress"`
}

type CinderSchedulerSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

// CinderStatus defines the observed state of Cinder
type CinderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cinder is the Schema for the cinders API
type Cinder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CinderSpec   `json:"spec,omitempty"`
	Status CinderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CinderList contains a list of Cinder
type CinderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cinder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cinder{}, &CinderList{})
}
