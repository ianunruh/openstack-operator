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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultVirtualHost = "openstack"
)

type ExternalBrokerAdminSecret struct {
	Name string `json:"name"`

	// +optional
	UsernameKey string `json:"usernameKey,omitempty"`

	// +optional
	PasswordKey string `json:"passwordKey,omitempty"`
}

type ExternalBrokerSpec struct {
	AdminSecret ExternalBrokerAdminSecret `json:"adminSecret"`

	Host string `json:"host"`

	// +optional
	Port uint16 `json:"port,omitempty"`

	// +optional
	AdminPort uint16 `json:"adminPort,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`
}

// RabbitMQUserSpec defines the desired state of RabbitMQUser
type RabbitMQUserSpec struct {
	Name        string `json:"name"`
	Secret      string `json:"secret"`
	VirtualHost string `json:"virtualHost"`

	// +optional
	Cluster string `json:"cluster,omitempty"`

	// +optional
	External *ExternalBrokerSpec `json:"external,omitempty"`

	// +optional
	SetupJob JobSpec `json:"setupJob,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`
}

func brokerDefault(spec RabbitMQUserSpec, instance, virtualHost string) RabbitMQUserSpec {
	if spec.Name == "" {
		spec.Name = instance
	}

	if spec.Secret == "" {
		spec.Secret = fmt.Sprintf("%s-rabbitmq", instance)
	}

	if spec.VirtualHost == "" {
		spec.VirtualHost = virtualHost
	}

	if spec.Cluster == "" && spec.External == nil {
		spec.Cluster = "rabbitmq"
	}

	return spec
}

// RabbitMQUserStatus defines the observed state of RabbitMQUser
type RabbitMQUserStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	SetupJobHash string `json:"setupJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.cluster`
//+kubebuilder:printcolumn:name="Vhost",type=string,JSONPath=`.spec.virtualHost`
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RabbitMQUser is the Schema for the rabbitmqusers API
type RabbitMQUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQUserSpec   `json:"spec,omitempty"`
	Status RabbitMQUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RabbitMQUserList contains a list of RabbitMQUser
type RabbitMQUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQUser{}, &RabbitMQUserList{})
}
