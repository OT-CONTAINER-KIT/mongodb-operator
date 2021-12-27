package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// KubernetesConfig will be the JSON struct for Basic MongoDB Config
type KubernetesConfig struct {
	Image            string                         `json:"image"`
	ImagePullPolicy  corev1.PullPolicy              `json:"imagePullPolicy,omitempty"`
	Resources        *corev1.ResourceRequirements   `json:"resources,omitempty"`
	ImagePullSecrets *[]corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// MongoDBSecurity is the JSON struct for MongoDB security configuration
type MongoDBSecurity struct {
	MongoDBAdminUser string                 `json:"mongoDBAdminUser,omitempty"`
	SecretRef        ExistingPasswordSecret `json:"secretRef,omitempty"`
}

// ExistingPasswordSecret is the struct to access the existing secret
type ExistingPasswordSecret struct {
	Name *string `json:"name,omitempty"`
	Key  *string `json:"key,omitempty"`
}

// Storage is the inteface to add pvc and pv support in MongoDB
type Storage struct {
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	StorageClassName *string                             `json:"storageClass,omitempty" protobuf:"bytes,5,opt,name=storageClassName"`
	StorageSize      string                              `json:"storageSize,omitempty" protobuf:"bytes,5,opt,name=storageClassName"`
}
