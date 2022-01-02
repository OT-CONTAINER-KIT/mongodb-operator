package k8sgo

import (
	"fmt"
	"github.com/thanhpk/randstr"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
)

const (
	KeyFileData = "kCiS/Cxw0lF2ZvHSqdtLGPsx77KwPf8e1hV0rXtViMVFlOgO2XVLsRqe12iRJT9gSUIQwo87tISyoyluO1EeZG6IpiZ+d/diwLzd3nRlIuDIvE/AnyrVN4HaWAMNeujK81z+vkmPvWj+g1C91l0k1iq93pCYTOS6POfdHIs/mcbWGs2WQggL2AXsTHjJTfBvD77Rm7nKrFu682zPs3xmUHkOADOigg+G4S4av2j6RvjCVjeUCuYVRL7/VPxFVY3iv/mCbSiijIxxsQalHbLAGlaQqVGJemtALcoyYTeoeCP20VnfZSxMl4QoCQP213Av4SRbpGZZpv2yZ2mhun5+e9+iyxUKQqAzLYPq/h0WYntQ+ZYs5qoaKwsLicox93o08W/S0hlhsTz4NGmWGIKfg064L+fIbDT2ep5lXcTH+z3MM+Rj7QdppVrBy1SNnMwMmzcv+f8ZRtIif10GxksDLwZgX66QhpmBfU+wxsD02TQVjnVtfGpGf4MkpRTOe+mPIY/ee4sUsNeBdg1P4iR2Sv49t1pTwwN4sWEXHcSDUSVCDpSEtBzN0A5V/+ONR5y5IlGcOjM9dAOErTWVngI82kJhyFFPRENdvmwrx7Sx57+DFcTk8xay0GCwhWgmaxEd/iW3ViJ8mo57hWegU7nss8ot1ro9/VJHHMQG+CeM/EWqTEFJuHBnebQliggqfez9sfqCdxS2rzdUQM6qwST+X4P3KEjhwD0iTCRHyVYel58ScExOkxdCuPTZyAv6MpYiCOa2CePh7fsXCiclrTBBamdri6YgRvXHFMuLsi3x6QswgJWWYwXXJywk6wGB/CeumPoDjVDCRVENsbxUCMWW3qeBEWzsHh74o9+nASVIfZalAT7DD7HORcVUkih/YMCDcR+iPc8SzvOpeDBe3zezXQTKNBC5BW63EP0xcuCfMkjVmwZYuA4DfeTFKqbG2u4bcy/W9J6jZHSqDMhk3sJNZX2d5wXar40SuqCVi3NifkWBiCBT"
)

// GenerateMongoKeyFile is a method to generate keyfile for mongodb cluster
func GenerateMongoKeyFile(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster-key")
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
		Password:    KeyFileData,
		Name:        appName,
	}
	err := CreateSecret(params)
	if err != nil {
		return err
	}
	return nil
}

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
	err := CreateOrUpdateStateFul(getMongoDBClusterParams(cr))
	if err != nil {
		logger.Error(err, "Cannot create cluster StatefulSet for MongoDB")
		return err
	}
	return nil
}

// CreateMongoClusterMonitoringSecret is a method to create secret for monitoring
func CreateMongoClusterMonitoringSecret(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "Secret")
	err := CreateSecret(getMongoDBClusterSecretParams(cr))
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
		Password:    password,
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
	containerArgs := []string{"mongod", "--bind_ip", "0.0.0.0", "--replSet", cr.ObjectMeta.Name, "--keyFile", "/mongodb-config/password"}
	params := statefulSetParameters{
		StatefulSetMeta: generateObjectMetaInformation(appName, cr.Namespace, labels, generateAnnotations()),
		OwnerDef:        mongoClusterAsOwner(cr),
		Namespace:       cr.Namespace,
		ContainerParams: containerParameters{
			Image:                 cr.Spec.KubernetesConfig.Image,
			ImagePullPolicy:       cr.Spec.KubernetesConfig.ImagePullPolicy,
			Resources:             cr.Spec.KubernetesConfig.Resources,
			MongoDBConatainerArgs: &containerArgs,
			ExtraVolumeMount:      getSecretVolumeMount(),
		},
		Replicas:     cr.Spec.MongoDBClusterSize,
		Labels:       labels,
		Annotations:  generateAnnotations(),
		ExtraVolumes: getSecretVolume(cr.ObjectMeta.Name),
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
