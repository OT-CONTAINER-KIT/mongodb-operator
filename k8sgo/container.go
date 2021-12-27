package k8sgo

import (
	corev1 "k8s.io/api/core/v1"
)

// containerParameters is the input struct for MongoDB container
type containerParameters struct {
	Image              string
	ImagePullPolicy    corev1.PullPolicy
	Resources          *corev1.ResourceRequirements
	PersistenceEnabled *bool
}

// generateContainerDef is to generate container definition for MongoDB
func generateContainerDef(name string, params containerParameters) []corev1.Container {
	return []corev1.Container{
		{
			Name:            "mongo",
			Image:           params.Image,
			ImagePullPolicy: params.ImagePullPolicy,
			VolumeMounts:    getVolumeMount(name, params.PersistenceEnabled),
		},
	}
}

// getVolumeMount is a method to create volume mounting list
func getVolumeMount(name string, persistenceEnabled *bool) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	if persistenceEnabled != nil && *persistenceEnabled {
		volumeMounts = []corev1.VolumeMount{
			{
				Name:      name,
				MountPath: "/data/db",
			},
		}
	}
	return volumeMounts
}
