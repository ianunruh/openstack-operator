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

// NovaFlavorSpec defines the desired state of NovaFlavor
type NovaFlavorSpec struct {
	//+optional
	Name string `json:"name,omitempty"`

	RAM   int  `json:"ram"`
	VCPUs int  `json:"vcpus"`
	Disk  *int `json:"disk"`

	//+optional
	Swap *int `json:"swap,omitempty"`

	//+optional
	Ephemeral *int `json:"ephemeral,omitempty"`

	//+optional
	IsPublic *bool `json:"isPublic,omitempty"`
}

// NovaFlavorStatus defines the observed state of NovaFlavor
type NovaFlavorStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	//+optional
	FlavorID string `json:"flavorID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// NovaFlavor is the Schema for the novaflavors API
type NovaFlavor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaFlavorSpec   `json:"spec,omitempty"`
	Status NovaFlavorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaFlavorList contains a list of NovaFlavor
type NovaFlavorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaFlavor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaFlavor{}, &NovaFlavorList{})
}
