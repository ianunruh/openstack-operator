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

// KeystoneServiceSpec defines the desired state of KeystoneService
type KeystoneServiceSpec struct {
	Name string `json:"name"`

	Type string `json:"type"`

	PublicURL   string `json:"publicURL"`
	InternalURL string `json:"internalURL"`
}

// KeystoneServiceStatus defines the observed state of KeystoneService
type KeystoneServiceStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	SetupJobHash string `json:"setupJobHash,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.publicURL`
// +kubebuilder:printcolumn:name="Internal URL",type=string,priority=1,JSONPath=`.spec.internalURL`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// KeystoneService is the Schema for the keystoneservices API
type KeystoneService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeystoneServiceSpec   `json:"spec,omitempty"`
	Status KeystoneServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeystoneServiceList contains a list of KeystoneService
type KeystoneServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeystoneService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeystoneService{}, &KeystoneServiceList{})
}
