package k8sgo

import (
	"fmt"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"mongodb-operator/mongo"
)

// InitializeMongoDBCluster is a method to create a mongodb cluster
func InitializeMongoDBCluster(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Cluster Setup")
	serviceName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	passwordParams := secretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name}
	password := getMongoDBPassword(passwordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongogo.MongoDBParameters{
		MongoURL:     mongoURL,
		Namespace:    cr.Namespace,
		Name:         cr.ObjectMeta.Name,
		ClusterNodes: cr.Spec.MongoDBClusterSize,
	}
	err := mongogo.InitiateMongoClusterRS(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create MongoDB cluster")
		return err
	}
	logger.Info("Successfully created the MongoDB cluster")
	return nil
}

// CheckMongoClusterState is a method to check mongodb cluster state
func CheckMongoClusterState(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Cluster Setup")
	serviceName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster")
	passwordParams := secretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name}
	password := getMongoDBPassword(passwordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongogo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
	}
	_, err := mongogo.CheckMongoClusterInitialized(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to check Mongo Cluster state")
		return err
	}
	logger.Info("Successfully checked the MongoDB cluster state")
	return nil
}

// CreateMongoDBMonitoringUser is a method to create a monitoring user for MongoDB
func CreateMongoDBMonitoringUser(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	serviceName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone")
	passwordParams := secretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name}
	password := getMongoDBPassword(passwordParams)
	monitoringPasswordParams := secretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: fmt.Sprintf("%s-%s", serviceName, "monitoring")}
	monitoringPassword := getMongoDBPassword(monitoringPasswordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongogo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		Password:  monitoringPassword,
	}
	err := mongogo.CreateMonitoringUser(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create monitoring user in MongoDB")
		return err
	}
	logger.Info("Successfully created the monitoring user")
	return nil
}

// CheckMonitoringUser is a method to check if monitoring user exists in MongoDB
func CheckMonitoringUser(cr *opstreelabsinv1alpha1.MongoDB) bool {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	serviceName := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone")
	passwordParams := secretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name}
	password := getMongoDBPassword(passwordParams)
	monitoringUser := "monitoring"
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongogo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		UserName:  &monitoringUser,
	}
	output, err := mongogo.GetMongoDBUser(mongoParams)
	if err != nil {
		return false
	}
	logger.Info("Successfully executed the command to check monitoring user")
	return output
}
