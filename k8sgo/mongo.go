package k8sgo

import (
	"fmt"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"mongodb-operator/mongo"
	"strings"
)

// InitializeMongoDBCluster is a method to create a mongodb cluster
func InitializeMongoDBCluster(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Cluster Setup")
	serviceName := fmt.Sprintf("%s-%s.%s", cr.ObjectMeta.Name, "cluster", cr.Namespace)
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongo.MongoDBParameters{
		MongoURL:     mongoURL,
		Namespace:    cr.Namespace,
		Name:         cr.ObjectMeta.Name,
		ClusterNodes: cr.Spec.MongoDBClusterSize,
		SetupType:    "standalone",
	}
	err := mongo.InitiateMongoClusterRS(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create MongoDB cluster")
		return err
	}
	logger.Info("Successfully created the MongoDB cluster")
	return nil
}

// CheckMongoClusterStateInitialized is a method to check mongodb cluster state
func CheckMongoClusterStateInitialized(cr *opstreelabsinv1alpha1.MongoDBCluster) (bool, error) {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Cluster Setup")
	serviceName := fmt.Sprintf("%s-%s.%s", cr.ObjectMeta.Name, "cluster", cr.Namespace)
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		SetupType: "standalone",
	}
	state, err := mongo.CheckMongoClusterInitialized(mongoParams)
	if err != nil {
		return state, err
	}
	logger.Info("Successfully checked the MongoDB cluster state")
	return state, nil
}

// CreateMongoDBMonitoringUser is a method to create a monitoring user for MongoDB
func CreateMongoDBMonitoringUser(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	serviceName := fmt.Sprintf("%s-%s.%s", cr.ObjectMeta.Name, "standalone", cr.Namespace)
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	monitoringPasswordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "standalone-monitoring"), SecretKey: "password"}
	monitoringPassword := GetMongoDBPassword(monitoringPasswordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		Password:  monitoringPassword,
		SetupType: "standalone",
	}
	err := mongo.CreateMonitoringUser(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create monitoring user in MongoDB")
		return err
	}
	logger.Info("Successfully created the monitoring user")
	return nil
}

// CreateMongoDBClusterMonitoringUser is a method to create a monitoring user for MongoDB
func CreateMongoDBClusterMonitoringUser(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	monitoringPasswordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, "cluster-monitoring"), SecretKey: "password"}
	monitoringPassword := GetMongoDBPassword(monitoringPasswordParams)
	mongoParams := mongo.MongoDBParameters{
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		Password:  monitoringPassword,
		SetupType: "cluster",
	}
	mongoURL := []string{"mongodb://", cr.Spec.MongoDBSecurity.MongoDBAdminUser, ":", password, "@"}
	for node := 0; node < int(*cr.Spec.MongoDBClusterSize); node++ {
		if node != int(*cr.Spec.MongoDBClusterSize) {
			mongoURL = append(mongoURL, fmt.Sprintf("%s,", mongo.GetMongoNodeInfo(mongoParams, node)))
		} else {
			mongoURL = append(mongoURL, mongo.GetMongoNodeInfo(mongoParams, node))
		}
	}
	mongoURL = append(mongoURL, fmt.Sprintf("/?replicaSet=%s", cr.ObjectMeta.Name))
	mongoParams.MongoURL = strings.Join(mongoURL, "")
	err := mongo.CreateMonitoringUser(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create monitoring user in MongoDB cluster")
		return err
	}
	logger.Info("Successfully created the monitoring user in MongoDB cluster")
	return nil
}

// CheckMongoDBClusterMonitoringUser is a method to check if monitoring user exists in MongoDB
func CheckMongoDBClusterMonitoringUser(cr *opstreelabsinv1alpha1.MongoDBCluster) bool {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	monitoringUser := "monitoring"
	mongoParams := mongo.MongoDBParameters{
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		UserName:  &monitoringUser,
		SetupType: "cluster",
	}
	mongoURL := []string{"mongodb://", cr.Spec.MongoDBSecurity.MongoDBAdminUser, ":", password, "@"}
	for node := 0; node < int(*cr.Spec.MongoDBClusterSize); node++ {
		if node != int(*cr.Spec.MongoDBClusterSize) {
			mongoURL = append(mongoURL, fmt.Sprintf("%s,", mongo.GetMongoNodeInfo(mongoParams, node)))
		} else {
			mongoURL = append(mongoURL, mongo.GetMongoNodeInfo(mongoParams, node))
		}
	}
	mongoURL = append(mongoURL, fmt.Sprintf("/?replicaSet=%s", cr.ObjectMeta.Name))
	mongoParams.MongoURL = strings.Join(mongoURL, "")
	output, err := mongo.GetMongoDBUser(mongoParams)
	if err != nil {
		return false
	}
	logger.Info("Successfully executed the command to check monitoring user in MongoDB cluster")
	return output
}

// CheckMonitoringUser is a method to check if monitoring user exists in MongoDB
func CheckMonitoringUser(cr *opstreelabsinv1alpha1.MongoDB) bool {
	logger := logGenerator(cr.ObjectMeta.Name, cr.Namespace, "MongoDB Monitoring User")
	serviceName := fmt.Sprintf("%s-%s.%s", cr.ObjectMeta.Name, "standalone", cr.Namespace) 
	passwordParams := SecretsParameters{Name: cr.ObjectMeta.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := GetMongoDBPassword(passwordParams)
	monitoringUser := "monitoring"
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.ObjectMeta.Name,
		UserName:  &monitoringUser,
		SetupType: "standalone",
	}
	output, err := mongo.GetMongoDBUser(mongoParams)
	if err != nil {
		return false
	}
	logger.Info("Successfully executed the command to check monitoring user")
	return output
}
