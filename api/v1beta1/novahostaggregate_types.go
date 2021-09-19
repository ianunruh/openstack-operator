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

// NovaHostAggregateSpec defines the desired state of NovaHostAggregate
type NovaHostAggregateSpec struct {
	//+optional
	Metadata map[string]string `json:"metadata,omitempty"`

	//+optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	//+optional
	Zone string `json:"zone,omitempty"`
}

// NovaHostAggregateStatus defines the observed state of NovaHostAggregate
type NovaHostAggregateStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	//+optional
	AggregateID int `json:"aggregateID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NovaHostAggregate is the Schema for the novahostaggregates API
type NovaHostAggregate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaHostAggregateSpec   `json:"spec,omitempty"`
	Status NovaHostAggregateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NovaHostAggregateList contains a list of NovaHostAggregate
type NovaHostAggregateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NovaHostAggregate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NovaHostAggregate{}, &NovaHostAggregateList{})
}
