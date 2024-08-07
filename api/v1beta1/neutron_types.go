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

// NeutronSpec defines the desired state of Neutron
type NeutronSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Server NeutronServerSpec `json:"server,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	MetadataAgent NeutronMetadataAgentSpec `json:"metadataAgent,omitempty"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`

	// +optional
	Cache CacheSpec `json:"cache,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`

	// +optional
	Nova NeutronNovaSpec `json:"nova,omitempty"`

	// +optional
	Placement NeutronPlacementSpec `json:"placement,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type NeutronServerSpec struct {
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

	// +optional
	TLSProxy TLSProxySpec `json:"tlsProxy,omitempty"`
}

type NeutronMetadataAgentSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type NeutronNovaSpec struct {
	// +optional
	MetadataHost string `json:"metadataHost,omitempty"`

	// +optional
	MetadataProtocol string `json:"metadataProtocol,omitempty"`

	// +optional
	Secret string `json:"secret,omitempty"`
}

type NeutronPlacementSpec struct {
	// +optional
	Secret string `json:"secret,omitempty"`
}

// NeutronStatus defines the observed state of Neutron
type NeutronStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Neutron is the Schema for the neutrons API
type Neutron struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NeutronSpec   `json:"spec,omitempty"`
	Status NeutronStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NeutronList contains a list of Neutron
type NeutronList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neutron `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Neutron{}, &NeutronList{})
}
