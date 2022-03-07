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

// NovaComputeNodeSpec defines the desired state of NovaComputeNode
type NovaComputeNodeSpec struct {
	Node string `json:"node"`

	Cell string `json:"cell"`
}

// NovaComputeNodeStatus defines the observed state of NovaComputeNode
type NovaComputeNodeStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	Hypervisor *NovaHypervisorStatus `json:"hypervisor,omitempty"`
}

type NovaHypervisorStatus struct {
	Enabled bool `json:"enabled"`

	Up bool `json:"up"`

	HostIP string `json:"hostIP"`

	HypervisorType string `json:"hypervisorType"`

	RunningServerCount int `json:"runningServerCount"`

	TaskCount int `json:"taskCount"`

	ServiceID string `json:"serviceID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// NovaComputeNode is the Schema for the novacomputenodes API
type NovaComputeNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaComputeNodeSpec   `json:"spec,omitempty"`
	Status NovaComputeNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaComputeNodeList contains a list of NovaComputeNode
type NovaComputeNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaComputeNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaComputeNode{}, &NovaComputeNodeList{})
}
