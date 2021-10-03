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

// NovaKeypairSpec defines the desired state of NovaKeypair
type NovaKeypairSpec struct {
	PublicKey string `json:"publicKey"`

	//+optional
	Name string `json:"name,omitempty"`

	//+optional
	User string `json:"user,omitempty"`

	//+optional
	UserDomain string `json:"userDomain,omitempty"`
}

// NovaKeypairStatus defines the observed state of NovaKeypair
type NovaKeypairStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// NovaKeypair is the Schema for the novakeypairs API
type NovaKeypair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaKeypairSpec   `json:"spec,omitempty"`
	Status NovaKeypairStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaKeypairList contains a list of NovaKeypair
type NovaKeypairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaKeypair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaKeypair{}, &NovaKeypairList{})
}
