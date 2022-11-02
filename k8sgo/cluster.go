package k8sgo

import (
	"fmt"
	"github.com/thanhpk/randstr"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
)

// CreateMongoClusterService is a method to create service for mongodb cluster
func CreateMongoClusterService(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Service")
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := serviceParameters{
		ServiceMeta:     generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoClusterAsOwner(cr),
		Namespace:       cr.Namespace,
		Labels:          labels,
		Annotations:     generateAnnotations(),
		HeadlessService: true,
		Port:            mongoDBPort,
		PortName:        "mongo",
	}
	err := CreateOrUpdateService(params)
	if err != nil {
		logger.Error(err, "Cannot create cluster Service for MongoDB")
		return err
	}
	return nil
}

// CreateMongoClusterMonitoringService is a method to create a monitoring service for mongodb cluster
func CreateMongoClusterMonitoringService(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Service")
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	monitoringParams := serviceParameters{
		ServiceMeta:     generateObjectMetaInformation(fmt.Sprintf("%s-%s", appName, "metrics"), cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoClusterAsOwner(cr),
		Namespace:       cr.Namespace,
		Labels:          labels,
		Annotations:     generateAnnotations(),
		HeadlessService: false,
		Port:            mongoDBMonitoringPort,
		PortName:        "metrics",
	}
	err := CreateOrUpdateService(monitoringParams)
	if err != nil {
		logger.Error(err, "Cannot create cluster metrics Service for MongoDB")
		return err
	}
	return nil
}

// CreateMongoClusterSetup is a method to create cluster statefulset for MongoDB
func CreateMongoClusterSetup(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "StatefulSet")
	err := CreateOrUpdateStateFul(getMongoDBClusterParams(cr), cr, nil)
	if err != nil {
		//logger.Error(err, "Cannot create cluster StatefulSet for MongoDB")
		return err
	}
	if cr.Spec.PodDisruptionBudget != nil && cr.Spec.PodDisruptionBudget.Enabled {
		err = CreateOrUpdatePodDisruption(getPodDisruptionParams(cr))
		if err != nil {
			logger.Error(err, "Cannot create PodDisruptionBudget for MongoDB")
			return err
		}
	}
	return nil
}

// CreateMongoClusterMonitoringSecret is a method to create secret for monitoring
func CreateMongoClusterMonitoringSecret(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Secret")
	err := CreateSecret(getMongoDBClusterSecretParams(cr), "password")
	if err != nil {
		logger.Error(err, "Cannot create mongodb monitoring secret for cluster")
		return err
	}
	return nil
}

// getMongoDBClusterSecretParams is a method to create secret for MongoDB Monitoring
func getMongoDBClusterSecretParams(cr *opstreelabsinv1alpha1.MongoDBCluster) secretsParameters {
	password := randstr.String(16)
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster-monitoring")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := secretsParameters{
		SecretsMeta: generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:    mongoClusterAsOwner(cr),
		Namespace:   cr.Namespace,
		Labels:      labels,
		Annotations: generateAnnotations(),
		Data:        password,
		Name:        appName,
	}
	return params
}

// getMongoDBClusterParams is a method to generate params for cluster
func getMongoDBClusterParams(cr *opstreelabsinv1alpha1.MongoDBCluster) statefulSetParameters {
	trueProperty := true
	falseProperty := false
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	monitoringSecretName := fmt.Sprintf("%s-%s", appName, "monitoring")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := statefulSetParameters{
		StatefulSetMeta: generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoClusterAsOwner(cr),
		Namespace:       cr.Namespace,
		ContainerParams: containerParameters{
			Image:               cr.Spec.KubernetesConfig.Image,
			ImagePullPolicy:     cr.Spec.KubernetesConfig.ImagePullPolicy,
			Resources:           cr.Spec.KubernetesConfig.Resources,
			MongoReplicaSetName: &cr.ObjectMeta.Name,
			MongoSetupType:      "cluster",
		},
		Replicas:          cr.Spec.MongoDBClusterSize,
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

	if cr.Spec.Security.TLS.Enabled {
		params.TLS = true
		params.ContainerParams.TLS = true
	}

	return params
}

// getPodDisruptionParams is a method to create parameters for pod disruption budget
func getPodDisruptionParams(cr *opstreelabsinv1alpha1.MongoDBCluster) PodDisruptionParameters {
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	labels := map[string]string{
		"app":           appName,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := PodDisruptionParameters{
		PDBMeta:        generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:       mongoClusterAsOwner(cr),
		Namespace:      cr.Namespace,
		Labels:         labels,
		MinAvailable:   cr.Spec.PodDisruptionBudget.MinAvailable,
		MaxUnavailable: cr.Spec.PodDisruptionBudget.MaxUnavailable,
	}
	return params
}
