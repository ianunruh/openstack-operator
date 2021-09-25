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

// KeystoneUserSpec defines the desired state of KeystoneUser
type KeystoneUserSpec struct {
	Secret string `json:"secret"`

	// +optional
	Roles []string `json:"roles,omitempty"`

	// +optional
	Domain string `json:"domain,omitempty"`

	// +optional
	Project string `json:"project,omitempty"`

	// +optional
	ProjectDomain string `json:"projectDomain,omitempty"`
}

// KeystoneUserStatus defines the observed state of KeystoneUser
type KeystoneUserStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	SetupJobHash string `json:"setupJobHash,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Project",type=string,JSONPath=`.spec.project`
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// KeystoneUser is the Schema for the keystoneusers API
type KeystoneUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeystoneUserSpec   `json:"spec,omitempty"`
	Status KeystoneUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeystoneUserList contains a list of KeystoneUser
type KeystoneUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeystoneUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeystoneUser{}, &KeystoneUserList{})
}
