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

// GlanceSpec defines the desired state of Glance
type GlanceSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	API GlanceAPISpec `json:"api,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Cache CacheSpec `json:"cache,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`

	// +optional
	Backends []GlanceBackendSpec `json:"backends,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type GlanceBackendSpec struct {
	Name string `json:"name"`

	// +optional
	Default bool `json:"default,omitempty"`

	// +optional
	Ceph *CephSpec `json:"ceph,omitempty"`

	// +optional
	PVC *VolumeSpec `json:"pvc,omitempty"`
}

type GlanceAPISpec struct {
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

// GlanceStatus defines the observed state of Glance
type GlanceStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Glance is the Schema for the glances API
type Glance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlanceSpec   `json:"spec,omitempty"`
	Status GlanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GlanceList contains a list of Glance
type GlanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Glance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Glance{}, &GlanceList{})
}
