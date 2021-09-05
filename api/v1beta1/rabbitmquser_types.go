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

// RabbitMQUserSpec defines the desired state of RabbitMQUser
type RabbitMQUserSpec struct {
	Cluster     string `json:"cluster"`
	Name        string `json:"name"`
	Secret      string `json:"secret"`
	VirtualHost string `json:"virtualHost"`
}

func brokerDefault(spec RabbitMQUserSpec, instance, virtualHost string) RabbitMQUserSpec {
	if spec.Cluster == "" {
		spec.Cluster = "rabbitmq"
	}

	if spec.Name == "" {
		spec.Name = instance
	}

	if spec.Secret == "" {
		spec.Secret = fmt.Sprintf("%s-rabbitmq", instance)
	}

	if spec.VirtualHost == "" {
		spec.VirtualHost = virtualHost
	}

	return spec
}

// RabbitMQUserStatus defines the observed state of RabbitMQUser
type RabbitMQUserStatus struct {
	Ready bool `json:"ready"`

	// +optional
	SetupJobHash string `json:"setupJobHash,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.cluster`
// +kubebuilder:printcolumn:name="Vhost",type=string,JSONPath=`.spec.virtualHost`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RabbitMQUser is the Schema for the rabbitmqusers API
type RabbitMQUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQUserSpec   `json:"spec,omitempty"`
	Status RabbitMQUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitMQUserList contains a list of RabbitMQUser
type RabbitMQUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQUser{}, &RabbitMQUserList{})
}
