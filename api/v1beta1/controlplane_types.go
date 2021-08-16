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

// ControlPlaneSpec defines the desired state of ControlPlane
type ControlPlaneSpec struct {
	Domain string `json:"domain"`

	// +optional
	Ingress ControlPlaneIngressSpec `json:"ingress"`

	Broker   RabbitMQSpec  `json:"broker"`
	Cache    MemcachedSpec `json:"cache"`
	Database MariaDBSpec   `json:"database"`

	Keystone  KeystoneSpec  `json:"keystone"`
	Glance    GlanceSpec    `json:"glance"`
	Placement PlacementSpec `json:"placement"`
	// +optional
	Cinder  CinderSpec  `json:"cinder"`
	Nova    NovaSpec    `json:"nova"`
	Neutron NeutronSpec `json:"neutron"`
	Horizon HorizonSpec `json:"horizon"`
	// +optional
	Heat HeatSpec `json:"heat"`
	// +optional
	Magnum MagnumSpec `json:"magnum"`
	// +optional
	Barbican BarbicanSpec `json:"barbican"`

	OVN OVNControlPlaneSpec `json:"ovn"`

	// +optional
	Octavia OctaviaSpec `json:"octavia"`
}

type ControlPlaneIngressSpec struct {
	// +optional
	Annotations map[string]string `json:"annotations"`

	// +optional
	TLSSecretName string `json:"tlsSecretName"`
}

// ControlPlaneStatus defines the observed state of ControlPlane
type ControlPlaneStatus struct {
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ControlPlane is the Schema for the controlplanes API
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ControlPlaneList contains a list of ControlPlane
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
