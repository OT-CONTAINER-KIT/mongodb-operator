package k8sgo

import (
	corev1 "k8s.io/api/core/v1"
)

// containerParameters is the input struct for MongoDB container
type containerParameters struct {
	Image                     string
	ImagePullPolicy           corev1.PullPolicy
	Resources                 *corev1.ResourceRequirements
	PersistenceEnabled        *bool
	MongoDBUser               *string
	SecretName                *string
	SecretKey                 *string
	MongoDBMonitoring         *bool
	MonitoringImage           string
	MonitoringImagePullPolicy *corev1.PullPolicy
	MonitoringSecret          *string
	MonitoringResources       *corev1.ResourceRequirements
}

// generateContainerDef is to generate container definition for MongoDB
func generateContainerDef(name string, params containerParameters) []corev1.Container {
	containerDef := []corev1.Container{
		{
			Name:            "mongo",
			Image:           params.Image,
			ImagePullPolicy: params.ImagePullPolicy,
			VolumeMounts:    getVolumeMount(name, params.PersistenceEnabled),
			Env:             getEnvironmentVariables(params),
		},
	}
	if params.Resources != nil {
		containerDef[0].Resources = *params.Resources
	}
	if params.MongoDBMonitoring != nil && *params.MongoDBMonitoring {
		containerDef = append(containerDef, getMongoDBExporterDef(params))
	}
	return containerDef
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

// getEnvironmentVariables is a method to create environment variables
func getEnvironmentVariables(params containerParameters) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	if params.SecretName != nil && params.MongoDBUser != nil {
		envVars = []corev1.EnvVar{
			{
				Name: "MONGO_INITDB_ROOT_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: *params.SecretName,
						},
						Key: *params.SecretKey,
					},
				},
			},
			{
				Name:  "MONGO_INITDB_ROOT_USERNAME",
				Value: *params.MongoDBUser,
			},
		}
	}
	return envVars
}

// getMongoDBExporterDef is a method to generate MongoDB Exporter
func getMongoDBExporterDef(params containerParameters) corev1.Container {
	containerDef := corev1.Container{
		Name:            "mongo-exporter",
		Image:           params.MonitoringImage,
		ImagePullPolicy: *params.MonitoringImagePullPolicy,
		Args:            []string{"--mongodb.uri=mongodb://$(MONGODB_MONITORING_USER):$(MONGODB_MONITORING_PASSWORD)@localhost:27017/admin"},
		Env: []corev1.EnvVar{
			{
				Name: "MONGODB_MONITORING_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: *params.MonitoringSecret,
						},
						Key: "password",
					},
				},
			},
			{
				Name:  "MONGODB_MONITORING_USER",
				Value: "monitoring",
			},
		},
	}

	if params.MonitoringResources != nil {
		containerDef.Resources = *params.MonitoringResources
	}
	return containerDef
}
