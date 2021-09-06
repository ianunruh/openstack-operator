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

// OVNControlPlaneSpec defines the desired state of OVNControlPlane
type OVNControlPlaneSpec struct {
	Image string `json:"image"`

	OVSDBNorth *OVSDBSpec `json:"ovsdbNorth"`
	OVSDBSouth *OVSDBSpec `json:"ovsdbSouth"`

	Node *OVNNodeSpec `json:"node"`

	// +optional
	Northd OVNNorthdSpec `json:"northd,omitempty"`
}

type OVSDBSpec struct {
	Volume *VolumeSpec `json:"volume"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type OVNNodeSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`

	// +optional
	BridgeMappings []string `json:"bridgeMappings,omitempty"`

	// +optional
	BridgePorts []string `json:"bridgePorts,omitempty"`
}

type OVNNorthdSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
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
