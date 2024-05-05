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

// OVNControlPlaneSpec defines the desired state of OVNControlPlane
type OVNControlPlaneSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	OVSDBNorth OVNDBSpec `json:"ovsdbNorth,omitempty"`

	// +optional
	OVSDBSouth OVNDBSpec `json:"ovsdbSouth,omitempty"`

	// +optional
	Controller OVNControllerSpec `json:"controller,omitempty"`

	// +optional
	Node OVNNodeSpec `json:"node,omitempty"`

	// +optional
	Northd OVNNorthdSpec `json:"northd,omitempty"`
}

type OVNDBSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Volume VolumeSpec `json:"volume"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OVNNodeSpec struct {
	// +optional
	DB OVSDBSpec `json:"db,omitempty"`

	// +optional
	Switch OVSSwitchSpec `json:"switch,omitempty"`

	// +optional
	OverlayCIDRs []string `json:"overlayCIDRs,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	BridgeMappings []string `json:"bridgeMappings,omitempty"`

	// +optional
	BridgePorts []string `json:"bridgePorts,omitempty"`
}

type OVNControllerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OVSSwitchSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OVSDBSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OVNNorthdSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// OVNControlPlaneStatus defines the observed state of OVNControlPlane
type OVNControlPlaneStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OVNControlPlane is the Schema for the ovncontrolplanes API
type OVNControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OVNControlPlaneSpec   `json:"spec,omitempty"`
	Status OVNControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OVNControlPlaneList contains a list of OVNControlPlane
type OVNControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OVNControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OVNControlPlane{}, &OVNControlPlaneList{})
}
