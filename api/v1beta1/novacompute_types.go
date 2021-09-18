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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NovaComputeSpec defines the desired state of NovaCompute
type NovaComputeSpec struct {
	Libvirtd NovaLibvirtdSpec `json:"libvirtd"`

	// +optional
	Cell string `json:"cell,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type NovaLibvirtdSpec struct {
	Image string `json:"image"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// NovaComputeStatus defines the observed state of NovaCompute
type NovaComputeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NovaCompute is the Schema for the novacomputes API
type NovaCompute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaComputeSpec   `json:"spec,omitempty"`
	Status NovaComputeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaComputeList contains a list of NovaCompute
type NovaComputeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaCompute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaCompute{}, &NovaComputeList{})
}
