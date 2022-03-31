package k8sgo

import (
	"fmt"
	"github.com/thanhpk/randstr"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
)

// CreateMongoStandaloneService is a method to create standalone service for MongoDB
func CreateMongoStandaloneService(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Service")
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "standalone",
		"role":          "standalone",
	}
	params := serviceParameters{
		ServiceMeta:     generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoAsOwner(cr),
		Namespace:       cr.Namespace,
		Labels:          labels,
		Annotations:     generateAnnotations(),
		HeadlessService: true,
		Port:            mongoDBPort,
		PortName:        "mongo",
	}
	err := CreateOrUpdateService(params)
	if err != nil {
		logger.Error(err, "Cannot create standalone Service for MongoDB")
		return err
	}
	monitoringParams := serviceParameters{
		ServiceMeta:     generateObjectMetaInformation(fmt.Sprintf("%s-%s", appName, "metrics"), cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoAsOwner(cr),
		Namespace:       cr.Namespace,
		Labels:          labels,
		Annotations:     generateAnnotations(),
		HeadlessService: false,
		Port:            mongoDBMonitoringPort,
		PortName:        "metrics",
	}
	err = CreateOrUpdateService(monitoringParams)
	if err != nil {
		logger.Error(err, "Cannot create standalone metrics Service for MongoDB")
		return err
	}
	return nil
}

// CreateMongoStandaloneSetup is a method to create standalone statefulset for MongoDB
func CreateMongoStandaloneSetup(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "StatefulSet")
	err := CreateOrUpdateStateFul(getMongoDBStandaloneParams(cr))
	if err != nil {
		logger.Error(err, "Cannot create standalone StatefulSet for MongoDB")
		return err
	}
	return nil
}

// CreateMongoMonitoringSecret is a method to create secret for monitoring
func CreateMongoMonitoringSecret(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Secret")
	err := CreateSecret(getMongoDBSecretParams(cr))
	if err != nil {
		logger.Error(err, "Cannot create mongodb monitoring secret")
		return err
	}
	return nil
}

// getMongoDBSecretParams is a method to create secret for MongoDB Monitoring
func getMongoDBSecretParams(cr *opstreelabsinv1alpha1.MongoDB) secretsParameters {
	password := randstr.String(16)
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone-monitoring")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "standalone",
		"role":          "standalone",
	}
	params := secretsParameters{
		SecretsMeta: generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:    mongoAsOwner(cr),
		Namespace:   cr.Namespace,
		Labels:      labels,
		Annotations: generateAnnotations(),
		Password:    password,
		Name:        appName,
	}
	return params
}

// getMongoDBStandaloneParams is a method to generate params for standalone
func getMongoDBStandaloneParams(cr *opstreelabsinv1alpha1.MongoDB) statefulSetParameters {
	replicas := int32(1)
	trueProperty := true
	falseProperty := false
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone")
	monitoringSecretName := fmt.Sprintf("%s-%s", appName, "monitoring")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "standalone",
		"role":          "standalone",
	}
	params := statefulSetParameters{
		StatefulSetMeta: generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoAsOwner(cr),
		Namespace:       cr.Namespace,
		ContainerParams: containerParameters{
			Image:           cr.Spec.KubernetesConfig.Image,
			ImagePullPolicy: cr.Spec.KubernetesConfig.ImagePullPolicy,
			Resources:       cr.Spec.KubernetesConfig.Resources,
			MongoSetupType:  "standalone",
		},
		Replicas:          &replicas,
		Labels:            labels,
		Annotations:       generateAnnotations(),
		NodeSelector:      cr.Spec.KubernetesConfig.NodeSelector,
		Affinity:          cr.Spec.KubernetesConfig.Affinity,
		PriorityClassName: cr.Spec.KubernetesConfig.PriorityClassName,
		Tolerations:       cr.Spec.KubernetesConfig.Tolerations,
		SecurityContext:   cr.Spec.KubernetesConfig.SecurityContext,
	}

	if cr.Spec.KubernetesConfig.ImagePullSecret != nil {
		params.ImagePullSecret = cr.Spec.KubernetesConfig.ImagePullSecret
	}
	if cr.Spec.MongoDBSecurity != nil {
		params.ContainerParams.MongoDBUser = &cr.Spec.MongoDBSecurity.MongoDBAdminUser
		params.ContainerParams.SecretName = cr.Spec.MongoDBSecurity.SecretRef.Name
		params.ContainerParams.SecretKey = cr.Spec.MongoDBSecurity.SecretRef.Key
	}
	if cr.Spec.MongoDBMonitoring != nil {
		params.ContainerParams.MongoDBMonitoring = &trueProperty
		params.ContainerParams.MonitoringSecret = &monitoringSecretName
		params.ContainerParams.MonitoringResources = cr.Spec.MongoDBMonitoring.Resources
		params.ContainerParams.MonitoringImage = cr.Spec.MongoDBMonitoring.Image
		params.ContainerParams.MonitoringImagePullPolicy = &cr.Spec.MongoDBMonitoring.ImagePullPolicy
	}
	if cr.Spec.MongoDBAdditionalConfig != nil {
		params.ContainerParams.AdditonalConfig = cr.Spec.MongoDBAdditionalConfig
		params.AdditionalConfig = cr.Spec.MongoDBAdditionalConfig
	}
	if cr.Spec.Storage != nil {
		params.ContainerParams.PersistenceEnabled = &trueProperty
		params.PVCParameters = pvcParameters{
			Name:             appName,
			Namespace:        cr.Namespace,
			Labels:           labels,
			Annotations:      generateAnnotations(),
			StorageSize:      cr.Spec.Storage.StorageSize,
			StorageClassName: cr.Spec.Storage.StorageClassName,
			AccessModes:      cr.Spec.Storage.AccessModes,
		}
	} else {
		params.ContainerParams.PersistenceEnabled = &falseProperty
	}
	return params
}
