package k8sgo

import (
	"fmt"
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
		logger.Error(err, "Cannot create standalone service for MongoDB")
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

func getMongoDBStandaloneParams(cr *opstreelabsinv1alpha1.MongoDB) statefulSetParameters {
	replicas := int32(1)
	trueProperty := true
	falseProperty := false
	appName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone")
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
		},
		Replicas:    &replicas,
		Labels:      labels,
		Annotations: generateAnnotations(),
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
