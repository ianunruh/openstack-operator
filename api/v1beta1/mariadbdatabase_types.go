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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExternalDatabaseAdminSecret struct {
	Name string `json:"name"`

	// +optional
	PasswordKey string `json:"passwordKey,omitempty"`
}

type ExternalDatabaseSpec struct {
	AdminSecret ExternalDatabaseAdminSecret `json:"adminSecret"`

	Host string `json:"host"`

	// +optional
	Port uint16 `json:"port,omitempty"`
}

// MariaDBDatabaseSpec defines the desired state of MariaDBDatabase
type MariaDBDatabaseSpec struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`

	// +optional
	Cluster string `json:"cluster,omitempty"`

	// +optional
	External *ExternalDatabaseSpec `json:"external,omitempty"`

	// +optional
	SetupJob JobSpec `json:"setupJob,omitempty"`
}

func databaseDefault(spec MariaDBDatabaseSpec, instance string) MariaDBDatabaseSpec {
	if spec.Name == "" {
		spec.Name = strings.ReplaceAll(instance, "-", "_")
	}

	if spec.Secret == "" {
		spec.Secret = fmt.Sprintf("%s-db", instance)
	}

	if spec.Cluster == "" && spec.External == nil {
		spec.Cluster = "mariadb"
	}

	return spec
}

// MariaDBDatabaseStatus defines the observed state of MariaDBDatabase
type MariaDBDatabaseStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	SetupJobHash string `json:"setupJobHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.cluster`
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MariaDBDatabase is the Schema for the mariadbdatabases API
type MariaDBDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBDatabaseSpec   `json:"spec,omitempty"`
	Status MariaDBDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MariaDBDatabaseList contains a list of MariaDBDatabase
type MariaDBDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBDatabase{}, &MariaDBDatabaseList{})
}
