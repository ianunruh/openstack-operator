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

// HorizonSpec defines the desired state of Horizon
type HorizonSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Server HorizonServerSpec `json:"server,omitempty"`

	// +optional
	SSO HorizonSSOSpec `json:"sso,omitempty"`

	// +optional
	Cache CacheSpec `json:"cache,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`
}

type HorizonServerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	TLS TLSServerSpec `json:"tls,omitempty"`
}

type HorizonSSOSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// +optional
	KeystoneURL string `json:"keystoneURL,omitempty"`

	// +optional
	Methods []HorizonSSOMethod `json:"methods,omitempty"`
}

type HorizonSSOMethod struct {
	Kind string `json:"kind"`

	Title string `json:"title"`

	// +optional
	Default bool `json:"default,omitempty"`
}

// HorizonStatus defines the observed state of Horizon
type HorizonStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Horizon is the Schema for the horizons API
type Horizon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizonSpec   `json:"spec,omitempty"`
	Status HorizonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HorizonList contains a list of Horizon
type HorizonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Horizon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Horizon{}, &HorizonList{})
}
