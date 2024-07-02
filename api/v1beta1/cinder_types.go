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

// CinderSpec defines the desired state of Cinder
type CinderSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	API CinderAPISpec `json:"api,omitempty"`

	// +optional
	Backup CinderBackupSpec `json:"backup,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	Scheduler CinderSchedulerSpec `json:"scheduler,omitempty"`

	// +optional
	Volume CinderVolumeSpec `json:"volume,omitempty"`

	Backends []CinderBackendSpec `json:"backends"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`

	// +optional
	Cache CacheSpec `json:"cache,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type CinderBackendSpec struct {
	Name string `json:"name"`

	VolumeBackendName string `json:"volumeBackendName"`

	// +optional
	Ceph *CephSpec `json:"ceph,omitempty"`
}

type CinderAPISpec struct {
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
}

type CinderBackupSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type CinderSchedulerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type CinderVolumeSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// CinderStatus defines the observed state of Cinder
type CinderStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Cinder is the Schema for the cinders API
type Cinder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CinderSpec   `json:"spec,omitempty"`
	Status CinderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CinderList contains a list of Cinder
type CinderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cinder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cinder{}, &CinderList{})
}
