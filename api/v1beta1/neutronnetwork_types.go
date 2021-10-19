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
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NeutronNetworkSpec defines the desired state of NeutronNetwork
type NeutronNetworkSpec struct {
	//+optional
	Name string `json:"name,omitempty"`

	//+optional
	Description string `json:"description,omitempty"`

	//+optional
	Shared *bool `json:"shared,omitempty"`

	//+optional
	External *bool `json:"external,omitempty"`

	//+optional
	AdminStateUp *bool `json:"adminStateUp,omitempty"`

	//+optional
	Project string `json:"project,omitempty"`

	//+optional
	AvailabilityZoneHints []string `json:"availabilityZoneHints,omitempty"`

	//+optional
	Segments []provider.Segment `json:"segments,omitempty"`
}

type NeutronNetworkSegment struct {
	NetworkType string `json:"networkType"`

	PhysicalNetwork string `json:"physicalNetwork"`

	//+optional
	SegmentationID int `json:"segmentationID,omitempty"`
}

// NeutronNetworkStatus defines the observed state of NeutronNetwork
type NeutronNetworkStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	//+optional
	ProviderID string `json:"providerID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// NeutronNetwork is the Schema for the neutronnetworks API
type NeutronNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NeutronNetworkSpec   `json:"spec,omitempty"`
	Status NeutronNetworkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NeutronNetworkList contains a list of NeutronNetwork
type NeutronNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NeutronNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NeutronNetwork{}, &NeutronNetworkList{})
}
