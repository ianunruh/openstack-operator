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

// NovaComputeNodeSpec defines the desired state of NovaComputeNode
type NovaComputeNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NovaComputeNode. Edit novacomputenode_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// NovaComputeNodeStatus defines the observed state of NovaComputeNode
type NovaComputeNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
