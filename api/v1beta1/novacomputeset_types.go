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

// NovaComputeSetSpec defines the desired state of NovaComputeSet
type NovaComputeSetSpec struct {
	Cell string `json:"cell,omitempty"`

	Image string `json:"image,omitempty"`

	Libvirtd NovaLibvirtdSpec `json:"libvirtd,omitempty"`

	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// NovaComputeSetStatus defines the observed state of NovaComputeSet
type NovaComputeSetStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NovaComputeSet is the Schema for the novacomputesets API
type NovaComputeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaComputeSetSpec   `json:"spec,omitempty"`
	Status NovaComputeSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaComputeSetList contains a list of NovaComputeSet
type NovaComputeSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaComputeSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaComputeSet{}, &NovaComputeSetList{})
}
