package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// KubernetesConfig will be the JSON struct for Basic MongoDB Config
type KubernetesConfig struct {
	Image                  string                         `json:"image"`
	ImagePullPolicy        corev1.PullPolicy              `json:"imagePullPolicy,omitempty"`
	Resources              *corev1.ResourceRequirements   `json:"resources,omitempty"`
	ExistingPasswordSecret *ExistingPasswordSecret        `json:"existingSecrets,omitempty"`
	ImagePullSecrets       *[]corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ExistingPasswordSecret is the struct to access the existing secret
type ExistingPasswordSecret struct {
	Name *string `json:"name,omitempty"`
	Key  *string `json:"key,omitempty"`
}

// Storage is the inteface to add pvc and pv support in MongoDB
type Storage struct {
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}
