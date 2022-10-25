package k8sgo

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// containerParameters is the input struct for MongoDB container
type containerParameters struct {
	Image                     string
	ImagePullPolicy           corev1.PullPolicy
	Resources                 *corev1.ResourceRequirements
	Command                   []string
	PersistenceEnabled        *bool
	MongoReplicaSetName       *string
	MongoSetupType            string
	MongoDBUser               *string
	SecretName                *string
	SecretKey                 *string
	MongoDBMonitoring         *bool
	MonitoringImage           string
	MonitoringImagePullPolicy *corev1.PullPolicy
	MonitoringSecret          *string
	MonitoringResources       *corev1.ResourceRequirements
	ExtraVolumeMount          *corev1.VolumeMount
	AdditonalConfig           *string
	TLS                       bool
}

// generateContainerDef is to generate container definition for MongoDB
func generateContainerDef(name string, params containerParameters) []corev1.Container {
	volumeMounts := getVolumeMount(name, params.PersistenceEnabled, params.AdditonalConfig, params.TLS)
	if params.ExtraVolumeMount != nil {
		volumeMounts = append(volumeMounts, *params.ExtraVolumeMount)
	}
	containerDef := []corev1.Container{
		{
			Name:            "mongo",
			Image:           params.Image,
			ImagePullPolicy: params.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			Command: []string{
				"mongod",
				"-f",
				"/etc/mongo.d/extra/mongo.yaml",
			},
			Env:            getEnvironmentVariables(params),
			ReadinessProbe: getMongoDBProbe(),
			LivenessProbe:  getMongoDBProbe(),
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
func getVolumeMount(name string, persistenceEnabled *bool, additionalConfig *string, tls bool) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	if persistenceEnabled != nil && *persistenceEnabled {
		volumeMounts = []corev1.VolumeMount{
			{
				Name:      name,
				MountPath: "/data/db",
			},
		}
	}

	if additionalConfig != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "external-config",
			MountPath: "/etc/mongo.d/extra",
		})
	}

	if tls {
		// mount ca volume
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      tlsCAVolumeName,
			MountPath: tlsCAMountPath,
		})
		// mount crt volume
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      tlsCertVolumeName,
			MountPath: tlsOperatorSecretMountPath,
		})
	}

	return volumeMounts
}

// getEnvironmentVariables is a method to create environment variables
func getEnvironmentVariables(params containerParameters) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	if params.SecretName != nil && params.MongoDBUser != nil {
		envVars = []corev1.EnvVar{
			{
				Name: "MONGO_ROOT_PASSWORD",
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
				Name:  "MONGO_ROOT_USERNAME",
				Value: *params.MongoDBUser,
			},
			{
				Name:  "MONGO_MODE",
				Value: params.MongoSetupType,
			},
		}
	}
	if params.MongoReplicaSetName != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MONGO_REPL",
			Value: *params.MongoReplicaSetName,
		})
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
		ReadinessProbe: getMonitoringProbe(),
		LivenessProbe:  getMonitoringProbe(),
	}

	if params.MonitoringResources != nil {
		containerDef.Resources = *params.MonitoringResources
	}
	return containerDef
}

// getMongoDBProbe is a method to generate probe info for MongoDB
func getMongoDBProbe() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: 15,
		PeriodSeconds:       15,
		FailureThreshold:    5,
		TimeoutSeconds:      5,
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"mongo",
					"--eval",
					"db.adminCommand('ping')",
				},
			},
		},
	}
}

// getMonitoringProbe is a method to generate probe info for Monitoring
func getMonitoringProbe() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: 15,
		PeriodSeconds:       15,
		FailureThreshold:    5,
		TimeoutSeconds:      5,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromInt(9216),
				Path: "/metrics",
			},
		},
	}
}
