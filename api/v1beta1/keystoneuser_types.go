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

// KeystoneUserSpec defines the desired state of KeystoneUser
type KeystoneUserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Secret string `json:"secret"`

	// +optional
	Roles []string `json:"roles"`

	// +optional
	Domain string `json:"domain"`

	// +optional
	Project string `json:"project"`

	// +optional
	ProjectDomain string `json:"projectDomain"`
}

// KeystoneUserStatus defines the observed state of KeystoneUser
type KeystoneUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
