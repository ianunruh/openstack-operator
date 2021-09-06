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

// ManilaSpec defines the desired state of Manila
type ManilaSpec struct {
	Image string `json:"image"`

	// +optional
	API ManilaAPISpec `json:"api,omitempty"`

	// +optional
	Scheduler ManilaSchedulerSpec `json:"scheduler,omitempty"`

	// +optional
	Share ManilaShareSpec `json:"share,omitempty"`

	Backends []ManilaBackendSpec `json:"backends"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`
}

type ManilaAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

type ManilaSchedulerSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type ManilaShareSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type ManilaBackendSpec struct {
	Name string `json:"name"`

	ShareBackendName string `json:"shareBackendName"`

	Ceph *ManilaCephSpec `json:"ceph"`
}

type ManilaCephSpec struct {
	ClientName string `json:"clientName"`

	Secret string `json:"secret"`

	Rook *ManilaRookCephSpec `json:"rook"`
}

type ManilaRookCephSpec struct {
	Namespace string `json:"namespace"`
}

// ManilaStatus defines the observed state of Manila
type ManilaStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Manila is the Schema for the manilas API
type Manila struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManilaSpec   `json:"spec,omitempty"`
	Status ManilaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManilaList contains a list of Manila
type ManilaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manila `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manila{}, &ManilaList{})
}
