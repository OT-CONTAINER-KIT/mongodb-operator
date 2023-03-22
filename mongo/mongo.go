package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_mongo")

const (
	dbName         = "admin"
	monitoringUser = "monitoring"
)

// MongoDBParameters is a struct for MongoDB related inputs
type MongoDBParameters struct {
	MongoURL     string
	SetupType    string
	Namespace    string
	Name         string
	Password     string
	UserName     *string
	ClusterNodes *int32
}

// InitiateMongoClient is a method to create client connection with MongoDB
func InitiateMongoClient(params MongoDBParameters) *mongo.Client {
	logger := logGenerator(params.Name, params.Namespace, "MongoDB Client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(params.MongoURL).SetDirect(true))
	if err != nil {
		logger.Error(err, "Unable to establish connection with MongoDB")
	}
	return client
}

// initiateMongoClusterClient is a method to create client connection with MongoDB Cluster
func initiateMongoClusterClient(params MongoDBParameters) *mongo.Client {
	logger := logGenerator(params.Name, params.Namespace, "MongoDB Cluster Client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(params.MongoURL))
	if err != nil {
		logger.Error(err, "Unable to establish connection with MongoDB Cluster")
	}
	return client
} 

// DiscconnectMongoClient is a method to disconnect MongoDB client
func DiscconnectMongoClient(client *mongo.Client) error {
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			return
		}
	}()
	return nil
}

// CreateMonitoringUser is a method to create monitoring user inside MongoDB
//
//nolint:govet
func CreateMonitoringUser(params MongoDBParameters) error {
	var client *mongo.Client
	if params.SetupType == "cluster" {
		client = initiateMongoClusterClient(params)
	} else {
		client = InitiateMongoClient(params)
	}
	response := client.Database(dbName).RunCommand(context.Background(), bson.D{
		{Key: "createUser", Value: monitoringUser}, {Key: "pwd", Value: params.Password},
		{Key: "roles", Value: []bson.M{{"role": "clusterMonitor", "db": "admin"}, {"role": "read", "db": "local"}}}},
	)

	if response.Err() != nil {
		return response.Err()
	}
	err := DiscconnectMongoClient(client)
	if err != nil {
		return err
	}
	return nil
}

// GetMongoDBUser is a method to check if user exists in MongoDB
//
//nolint:govet
func GetMongoDBUser(params MongoDBParameters) (bool, error) {
	var client *mongo.Client
	if params.SetupType == "cluster" {
		client = initiateMongoClusterClient(params)
	} else {
		client = InitiateMongoClient(params)
	}
	collection := client.Database("admin").Collection("system.users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Count().SetMaxTime(2 * time.Second)
	docsCount, err := collection.CountDocuments(ctx, bson.D{{Key: "user", Value: *params.UserName}}, opts)
	if err != nil {
		return false, err
	}
	err = DiscconnectMongoClient(client)
	if err != nil {
		return false, err
	}
	if docsCount > 0 {
		return true, nil
	}
	return false, nil
}

// InitiateMongoClusterRS is a method to create MongoDB cluster
func InitiateMongoClusterRS(params MongoDBParameters) error {
	var mongoNodeInfo []bson.M
	client := InitiateMongoClient(params)
	for node := 0; node < int(*params.ClusterNodes); node++ {
		mongoNodeInfo = append(mongoNodeInfo, bson.M{"_id": node, "host": GetMongoNodeInfo(params, node)})
	}
	config := bson.M{
		"_id":     params.Name,
		"members": mongoNodeInfo,
	}
	// command in doc
	response := client.Database(dbName).RunCommand(context.Background(), bson.M{"replSetInitiate": config})
	if response.Err() != nil {
		return response.Err()
	}
	err := DiscconnectMongoClient(client)
	if err != nil {
		return err
	}
	return nil
}

// CheckMongoClusterInitialized is a method to check if cluster is initailized or not
func CheckMongoClusterInitialized(params MongoDBParameters) (bool, error) {
	client := InitiateMongoClient(params)
	var result bson.M
	err := client.Database(dbName).RunCommand(context.Background(), bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&result)
	if err != nil {
		return false, err
	}
	if result["ok"] != 0 {
		return true, nil
	}
	return false, nil
}

// GetMongoNodeInfo is a method to get info for MongoDB node
func GetMongoNodeInfo(params MongoDBParameters, count int) string {
	return fmt.Sprintf("%s-cluster-%v.%s-cluster.%s:27017", params.Name, count, params.Name, params.Namespace)
}

// logGenerator is a method to generate logging interface
func logGenerator(name, namespace, resourceType string) logr.Logger {
	reqLogger := log.WithValues("Namespace", namespace, "Name", name, "Resource Type", resourceType)
	return reqLogger
}
