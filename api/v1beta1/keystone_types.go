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

// KeystoneSpec defines the desired state of Keystone
type KeystoneSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	API KeystoneAPISpec `json:"api,omitempty"`

	// +optional
	BootstrapJob JobSpec `json:"bootstrapJob,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	Database MariaDBDatabaseSpec `json:"database,omitempty"`

	// +optional
	Broker RabbitMQUserSpec `json:"broker,omitempty"`

	// +optional
	Cache CacheSpec `json:"cache,omitempty"`

	// +optional
	TLS TLSClientSpec `json:"tls,omitempty"`

	// +optional
	Notifications KeystoneNotificationsSpec `json:"notifications,omitempty"`

	// +optional
	OIDC KeystoneOIDCSpec `json:"oidc,omitempty"`

	// +optional
	ExtraConfig ExtraConfig `json:"extraConfig,omitempty"`
}

type KeystoneAPISpec struct {
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
}

type KeystoneNotificationsSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

type KeystoneOIDCSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// +optional
	Secret string `json:"secret,omitempty"`

	// +optional
	IdentityProvider string `json:"identityProvider,omitempty"`

	// +optional
	DashboardURL string `json:"dashboardURL,omitempty"`

	// +optional
	ProviderMetadataURL string `json:"providerMetadataURL,omitempty"`

	// +optional
	RedirectURI string `json:"redirectURI,omitempty"`

	// +optional
	Scopes []string `json:"scopes,omitempty"`

	// +optional
	RequireClaims []string `json:"requireClaims,omitempty"`

	// +optional
	ExtraConfig map[string]string `json:"extraConfig,omitempty"`
}

// KeystoneStatus defines the observed state of Keystone
type KeystoneStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	BootstrapJobHash string `json:"bootstrapJobHash,omitempty"`

	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Keystone is the Schema for the keystones API
type Keystone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeystoneSpec   `json:"spec,omitempty"`
	Status KeystoneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KeystoneList contains a list of Keystone
type KeystoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Keystone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Keystone{}, &KeystoneList{})
}
