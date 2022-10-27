package k8sgo

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"mongodb-operator/mongo"
	"strings"
)

// CheckMongoClusterState is a method to check mongodb cluster state
func CheckMongoClusterState(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.Name, cr.Namespace, "MongoDB Cluster Setup")
	passwordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := getMongoDBPassword(passwordParams)
	username := cr.Spec.MongoDBSecurity.MongoDBAdminUser
	mongoParams := mongogo.MongoDBParameters{
		Namespace:    cr.Namespace,
		Name:         cr.Name,
		ClusterNodes: cr.Spec.MongoDBClusterSize,
		SetupType:    "cluster",
		Password:     password,
		UserName:     &username,
	}
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s/", username, password, mongogo.GetMongoNodeInfo(mongoParams, 0))
	mongoParams.MongoURL = mongoURL
	err, result := mongogo.CheckReplSetGetStatus(mongoParams)
	if err != nil {
		return err
	}

	if result != nil || result["members"] != "" {
		members := result["members"].(primitive.A)
		membersArray := []interface{}(members)
		if int(*mongoParams.ClusterNodes) != len(membersArray) {
			var version int64
			version = result["term"].(int64)
			err := GetMongoDBParamsForScaling(mongoParams, len(membersArray), int(version))
			if err != nil {
				return err
			}
		}
	}

	logger.Info("Successfully checked the MongoDB cluster state")
	return nil
}

// CreateMongoDBMonitoringUser is a method to create a monitoring user for MongoDB
func CreateMongoDBMonitoringUser(cr *opstreelabsinv1alpha1.MongoDB) error {
	logger := logGenerator(cr.Name, cr.Namespace, "MongoDB Monitoring User")
	serviceName := fmt.Sprintf("%s-%s.%s", cr.Name, "standalone", cr.Namespace)
	passwordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := getMongoDBPassword(passwordParams)
	monitoringPasswordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: fmt.Sprintf("%s-%s", cr.Name, "standalone-monitoring"), SecretKey: "password"}
	monitoringPassword := getMongoDBPassword(monitoringPasswordParams)
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", cr.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)
	mongoParams := mongogo.MongoDBParameters{
		MongoURL:  mongoURL,
		Namespace: cr.Namespace,
		Name:      cr.Name,
		Password:  monitoringPassword,
		SetupType: "standalone",
	}
	err := mongogo.CreateMonitoringUser(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create monitoring user in MongoDB")
		return err
	}
	logger.Info("Successfully created the monitoring user")
	return nil
}

// CreateMongoDBClusterMonitoringUser is a method to create a monitoring user for MongoDB
func CreateMongoDBClusterMonitoringUser(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.Name, cr.Namespace, "MongoDB Monitoring User")
	passwordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := getMongoDBPassword(passwordParams)
	monitoringPasswordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: fmt.Sprintf("%s-%s", cr.Name, "cluster-monitoring"), SecretKey: "password"}
	monitoringPassword := getMongoDBPassword(monitoringPasswordParams)
	mongoParams := mongogo.MongoDBParameters{
		Namespace: cr.Namespace,
		Name:      cr.Name,
		Password:  monitoringPassword,
		SetupType: "cluster",
	}
	mongoURL := []string{"mongodb://", cr.Spec.MongoDBSecurity.MongoDBAdminUser, ":", password, "@"}
	for node := 0; node < int(*cr.Spec.MongoDBClusterSize); node++ {
		if node != int(*cr.Spec.MongoDBClusterSize) {
			mongoURL = append(mongoURL, fmt.Sprintf("%s,", mongogo.GetMongoNodeInfo(mongoParams, node)))
		} else {
			mongoURL = append(mongoURL, mongogo.GetMongoNodeInfo(mongoParams, node))
		}
	}
	mongoURL = append(mongoURL, fmt.Sprintf("/?replicaSet=%s", cr.Name))
	mongoParams.MongoURL = strings.Join(mongoURL, "")
	err := mongogo.CreateMonitoringUser(mongoParams)
	if err != nil {
		logger.Error(err, "Unable to create monitoring user in MongoDB cluster")
		return err
	}
	logger.Info("Successfully created the monitoring user in MongoDB cluster")
	return nil
}

// CheckMongoDBClusterMonitoringUser is a method to check if monitoring user exists in MongoDB
func CheckMongoDBClusterMonitoringUser(cr *opstreelabsinv1alpha1.MongoDBCluster) error {
	logger := logGenerator(cr.Name, cr.Namespace, "MongoDB Monitoring User")
	passwordParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: *cr.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *cr.Spec.MongoDBSecurity.SecretRef.Key}
	password := getMongoDBPassword(passwordParams)
	monitorSecretName := fmt.Sprintf("%s-cluster-monitoring", cr.ObjectMeta.Name)
	monitoringPassParams := secretsParameters{Name: cr.Name, Namespace: cr.Namespace, SecretName: monitorSecretName, SecretKey: "password"}
	monitoringUser := "monitoring"
	monitoringPass := getMongoDBPassword(monitoringPassParams)
	mongoParams := mongogo.MongoDBParameters{
		Namespace: cr.Namespace,
		Name:      cr.Name,
		UserName:  &monitoringUser,
		SetupType: "cluster",
		Password:  monitoringPass,
	}
	err := mongogo.GetOrCreateMonitoringUser(GetMongoDBClusterURL(mongoParams, cr, password))
	if err != nil {
		return err
	}
	logger.Info("Successfully executed the command to check monitoring user in MongoDB cluster")
	return nil
}

func GetMongoDBClusterURL(mongoParams mongogo.MongoDBParameters, cr *opstreelabsinv1alpha1.MongoDBCluster, password string) mongogo.MongoDBParameters {
	mongoClusterURL := []string{"mongodb://", cr.Spec.MongoDBSecurity.MongoDBAdminUser, ":", password, "@"}
	for node := 0; node < int(*cr.Spec.MongoDBClusterSize); node++ {
		if node != int(*cr.Spec.MongoDBClusterSize) {
			mongoClusterURL = append(mongoClusterURL, fmt.Sprintf("%s,", mongogo.GetMongoNodeInfo(mongoParams, node)))
		} else {
			mongoClusterURL = append(mongoClusterURL, mongogo.GetMongoNodeInfo(mongoParams, node))
		}
	}
	mongoClusterURL = append(mongoClusterURL, fmt.Sprintf("/?replicaSet=%s", cr.Name))
	mongoParams.MongoClusterURL = strings.Join(mongoClusterURL, "")
	return mongoParams
}

func GetMongoDBParamsForScaling(mongoParams mongogo.MongoDBParameters, members int, version int) error {
	var mongoClusterURL string
	for node := 0; node < members; node++ {
		mongoClusterURL += mongogo.GetMongoNodeInfo(mongoParams, node) + ","
	}
	mongoClusterURL = strings.TrimRight(mongoClusterURL, ",")
	mongoClusterURL = fmt.Sprintf("mongodb://%s:%s@%s/?replicaSet=%s", *mongoParams.UserName, mongoParams.Password, mongoClusterURL, mongoParams.Name)
	mongoParams.MongoClusterURL = mongoClusterURL
	mongoParams.Version = version + 1
	err := mongogo.ScalingMongoClusterRS(mongoParams)
	if err != nil {
		Log.Error(err, "Unable to Scaling MongoDB cluster")
		return err
	}
	Log.Info("Successfully Scaling MongoDB cluster")
	return nil

}
