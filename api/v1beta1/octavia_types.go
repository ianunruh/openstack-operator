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

// OctaviaSpec defines the desired state of Octavia
type OctaviaSpec struct {
	// deprecated, use component specific images instead
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Amphora OctaviaAmphoraSpec `json:"amphora,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	OVN OctaviaOVNSpec `json:"ovn,omitempty"`

	// +optional
	API OctaviaAPISpec `json:"api,omitempty"`

	// +optional
	DBSyncJob JobSpec `json:"dbSyncJob,omitempty"`

	// +optional
	DriverAgent OctaviaDriverAgentSpec `json:"driverAgent,omitempty"`

	// +optional
	HealthManager OctaviaHealthManagerSpec `json:"healthManager,omitempty"`

	// +optional
	Housekeeping OctaviaHousekeepingSpec `json:"housekeeping,omitempty"`

	// +optional
	Worker OctaviaWorkerSpec `json:"worker,omitempty"`

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

type OctaviaAmphoraSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// +optional
	ImageURL string `json:"imageURL,omitempty"`

	// +optional
	ManagementCIDR string `json:"managementCIDR,omitempty"`
}

type OctaviaOVNSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

type OctaviaAPISpec struct {
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

type OctaviaDriverAgentSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// deprecated, driver agent has been merged with API
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// deprecated, driver agent has been merged with API
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OctaviaHealthManagerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	InitDHCPImage string `json:"initDHCPImage,omitempty"`

	// +optional
	InitOVSImage string `json:"initOVSImage,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OctaviaHousekeepingSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OctaviaWorkerSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// OctaviaStatus defines the observed state of Octavia
type OctaviaStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	DBSyncJobHash string `json:"dbSyncJobHash,omitempty"`

	// +optional
	Amphora OctaviaAmphoraStatus `json:"amphora,omitempty"`
}

type OctaviaAmphoraStatus struct {
	// +optional
	FlavorID string `json:"flavorID,omitempty"`

	// +optional
	ImageProjectID string `json:"imageProjectID,omitempty"`

	// +optional
	NetworkIDs []string `json:"networkIDs,omitempty"`

	// +optional
	SecurityGroupIDs []string `json:"securityGroupIDs,omitempty"`

	// +optional
	HealthPorts []OctaviaAmphoraHealthPort `json:"healthPorts,omitempty"`

	// +optional
	HealthSecurityGroupIDs []string `json:"healthSecurityGroupIDs,omitempty"`
}

type OctaviaAmphoraHealthPort struct {
	ID string `json:"id"`

	// +optional
	Name string `json:"name,omitempty"`

	MACAddress string `json:"macAddress"`

	IPAddress string `json:"ipAddress"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=octavias
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Octavia is the Schema for the octavia API
type Octavia struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OctaviaSpec   `json:"spec,omitempty"`
	Status OctaviaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OctaviaList contains a list of Octavia
type OctaviaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Octavia `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Octavia{}, &OctaviaList{})
}
