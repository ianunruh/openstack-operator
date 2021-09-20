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

// RabbitMQSpec defines the desired state of RabbitMQ
type RabbitMQSpec struct {
	Image string `json:"image"`

	Volume *VolumeSpec `json:"volume"`

	// +optional
	Management RabbitMQManagementSpec `json:"management,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Prometheus RabbitMQPrometheusSpec `json:"prometheus,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type RabbitMQManagementSpec struct {
	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

type RabbitMQPrometheusSpec struct {
	// +optional
	ServiceMonitor bool `json:"serviceMonitor,omitempty"`
}

// RabbitMQStatus defines the observed state of RabbitMQ
type RabbitMQStatus struct {
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RabbitMQ is the Schema for the rabbitmqs API
type RabbitMQ struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSpec   `json:"spec,omitempty"`
	Status RabbitMQStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitMQList contains a list of RabbitMQ
type RabbitMQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQ `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQ{}, &RabbitMQList{})
}
