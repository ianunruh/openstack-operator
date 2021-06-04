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

// NeutronSpec defines the desired state of Neutron
type NeutronSpec struct {
	Image string `json:"image"`

	// +optional
	Server NeutronServerSpec `json:"server"`

	DHCPAgent        NeutronDHCPAgentSpec        `json:"dhcpAgent"`
	L3Agent          NeutronL3AgentSpec          `json:"l3Agent"`
	LinuxBridgeAgent NeutronLinuxBridgeAgentSpec `json:"linuxBridgeAgent"`
	MetadataAgent    NeutronMetadataAgentSpec    `json:"metadataAgent"`

	Database MariaDBDatabaseSpec `json:"database"`

	Broker RabbitMQUserSpec `json:"broker"`
}

type NeutronServerSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

type NeutronLinuxBridgeAgentSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`

	// +optional
	PhysicalInterfaceMappings []string `json:"physicalInterfaceMappings"`
}

type NeutronDHCPAgentSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`
}

type NeutronL3AgentSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`
}

type NeutronMetadataAgentSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`
}

// NeutronStatus defines the observed state of Neutron
type NeutronStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Neutron is the Schema for the neutrons API
type Neutron struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NeutronSpec   `json:"spec,omitempty"`
	Status NeutronStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NeutronList contains a list of Neutron
type NeutronList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neutron `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Neutron{}, &NeutronList{})
}
