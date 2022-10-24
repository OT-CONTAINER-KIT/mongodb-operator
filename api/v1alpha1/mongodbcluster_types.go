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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Phase string

// MongoDBClusterSpec defines the desired state of MongoDBCluster
type MongoDBClusterSpec struct {
	MongoDBClusterSize      *int32                      `json:"clusterSize"`
	EnableArbiter           *bool                       `json:"enableMongoArbiter,omitempty"`
	KubernetesConfig        KubernetesConfig            `json:"kubernetesConfig"`
	Storage                 *Storage                    `json:"storage,omitempty"`
	MongoDBSecurity         *MongoDBSecurity            `json:"mongoDBSecurity"`
	MongoDBMonitoring       *MongoDBMonitoring          `json:"mongoDBMonitoring,omitempty"`
	PodDisruptionBudget     *MongoDBPodDisruptionBudget `json:"podDisruptionBudget,omitempty"`
	MongoDBAdditionalConfig *string                     `json:"mongoDBAdditionalConfig,omitempty"`
	Security                Security                    `json:"security"`
}

type Security struct {
	TLS TLS `json:"tls"`
}

// TLS is the configuration used to set up TLS encryption
type TLS struct {
	Enabled bool `json:"enabled"`

	// Optional configures if TLS should be required or optional for connections
	// +optional
	Optional bool `json:"optional,omitempty"`

	// CertificateKeySecret is a reference to a Secret containing a private key and certificate to use for TLS.
	// The key and cert are expected to be PEM encoded and available at "tls.key" and "tls.crt".
	// This is the same format used for the standard "kubernetes.io/tls" Secret type, but no specific type is required.
	// Alternatively, an entry tls.pem, containing the concatenation of cert and key, can be provided.
	// If all of tls.pem, tls.crt and tls.key are present, the tls.pem one needs to be equal to the concatenation of tls.crt and tls.key
	// +optional
	CertificateKeySecret LocalObjectReference `json:"certificateKeySecretRef,omitempty"`

	// CaCertificateSecret is a reference to a Secret containing the certificate for the CA which signed the server certificates
	// The certificate is expected to be available under the key "ca.crt"
	// +optional
	CaCertificateSecret *LocalObjectReference `json:"caCertificateSecretRef,omitempty"`

	// CaConfigMap is a reference to a ConfigMap containing the certificate for the CA which signed the server certificates
	// The certificate is expected to be available under the key "ca.crt"
	// This field is ignored when CaCertificateSecretRef is configured
	// +optional
	CaConfigMap *LocalObjectReference `json:"caConfigMapRef,omitempty"`
}

type LocalObjectReference struct {
	Name string `json:"name"`
}

// MongoDBPodDisruptionBudget defines the struct for MongoDB cluster
type MongoDBPodDisruptionBudget struct {
	Enabled        bool   `json:"enabled,omitempty"`
	MinAvailable   *int32 `json:"minAvailable,omitempty"`
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

// MongoDBClusterStatus defines the observed state of MongoDBCluster
type MongoDBClusterStatus struct {
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongoDBCluster is the Schema for the mongodbclusters API
type MongoDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBClusterSpec   `json:"spec,omitempty"`
	Status MongoDBClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoDBClusterList contains a list of MongoDBCluster
type MongoDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBCluster{}, &MongoDBClusterList{})
}
