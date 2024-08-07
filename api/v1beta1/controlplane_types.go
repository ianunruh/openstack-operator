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
	Ingress ControlPlaneIngressSpec `json:"ingress,omitempty"`

	// +optional
	NodeSelector ControlPlaneNodeSelector `json:"nodeSelector,omitempty"`

	// +optional
	TLS ControlPlaneTLSSpec `json:"tls,omitempty"`

	// +optional
	Broker RabbitMQSpec `json:"broker,omitempty"`

	// +optional
	ExternalBroker *ExternalBrokerSpec `json:"externalBroker,omitempty"`

	// +optional
	Cache MemcachedSpec `json:"cache,omitempty"`

	// +optional
	ExternalCache *CacheSpec `json:"externalCache,omitempty"`

	// +optional
	Database MariaDBSpec `json:"database,omitempty"`

	// +optional
	ExternalDatabase *ExternalDatabaseSpec `json:"externalDatabase,omitempty"`

	// +optional
	Keystone KeystoneSpec `json:"keystone,omitempty"`

	// +optional
	Glance GlanceSpec `json:"glance,omitempty"`

	// +optional
	Placement PlacementSpec `json:"placement,omitempty"`

	// +optional
	Nova NovaSpec `json:"nova,omitempty"`

	// +optional
	Neutron NeutronSpec `json:"neutron,omitempty"`

	// +optional
	OVN OVNControlPlaneSpec `json:"ovn,omitempty"`

	// +optional
	Horizon HorizonSpec `json:"horizon,omitempty"`

	// +optional
	Barbican *BarbicanSpec `json:"barbican,omitempty"`

	// +optional
	Cinder *CinderSpec `json:"cinder,omitempty"`

	// +optional
	Heat *HeatSpec `json:"heat,omitempty"`

	// +optional
	Magnum *MagnumSpec `json:"magnum,omitempty"`

	// +optional
	Manila *ManilaSpec `json:"manila,omitempty"`

	// +optional
	Octavia *OctaviaSpec `json:"octavia,omitempty"`

	// +optional
	Rally *RallySpec `json:"rally,omitempty"`
}

type ControlPlaneNodeSelector struct {
	// +optional
	Controller map[string]string `json:"controller,omitempty"`

	// +optional
	Compute map[string]string `json:"compute,omitempty"`
}

type ControlPlaneIngressSpec struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +optional
	ClassName *string `json:"className,omitempty"`

	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

type ControlPlaneTLSSpec struct {
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// +optional
	Client TLSClientSpec `json:"client,omitempty"`

	// +optional
	Server TLSServerSpec `json:"server,omitempty"`

	// +optional
	ExternalIssuer bool `json:"externalIssuer,omitempty"`
}

// ControlPlaneStatus defines the observed state of ControlPlane
type ControlPlaneStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ControlPlane is the Schema for the controlplanes API
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ControlPlaneList contains a list of ControlPlane
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
