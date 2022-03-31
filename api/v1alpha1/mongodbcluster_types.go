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
}

// MongoDBPodDisruptionBudget defines the struct for MongoDB cluster
type MongoDBPodDisruptionBudget struct {
	Enabled        bool   `json:"enabled,omitempty"`
	MinAvailable   *int32 `json:"minAvailable,omitempty"`
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

// MongoDBClusterStatus defines the observed state of MongoDBCluster
type MongoDBClusterStatus struct {
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
