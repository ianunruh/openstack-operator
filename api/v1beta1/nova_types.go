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

// NovaSpec defines the desired state of Nova
type NovaSpec struct {
	Image string `json:"image"`

	// +optional
	API NovaAPISpec `json:"api"`

	// +optional
	Conductor NovaConductorSpec `json:"conductor"`

	// +optional
	Scheduler NovaSchedulerSpec `json:"scheduler"`

	Libvirtd NovaLibvirtdSpec `json:"libvirtd"`

	Compute NovaComputeSpec `json:"compute"`

	APIDatabase MariaDBDatabaseSpec `json:"apiDatabase"`

	CellDatabase MariaDBDatabaseSpec `json:"cellDatabase"`

	Broker RabbitMQUserSpec `json:"broker"`

	Cells []NovaCellSpec `json:"cells"`
}

type NovaAPISpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

type NovaConductorSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

type NovaMetadataSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

type NovaNoVNCProxySpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Ingress *IngressSpec `json:"ingress"`
}

type NovaSchedulerSpec struct {
	// +optional
	Replicas int32 `json:"replicas"`
}

type NovaLibvirtdSpec struct {
	Image string `json:"image"`
}

type NovaComputeSpec struct {
	NodeSelector map[string]string `json:"nodeSelector"`
}

// NovaStatus defines the observed state of Nova
type NovaStatus struct {
	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=novas

// Nova is the Schema for the nova API
type Nova struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NovaSpec   `json:"spec,omitempty"`
	Status NovaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NovaList contains a list of Nova
type NovaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nova `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nova{}, &NovaList{})
}
