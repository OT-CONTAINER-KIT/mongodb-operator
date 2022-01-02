package mongogo

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var log = logf.Log.WithName("controller_mongo")

const (
	dbName         = "admin"
	monitoringUser = "monitoring"
)

// MongoDBParameters is a struct for MongoDB related inputs
type MongoDBParameters struct {
	MongoURL     string
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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(params.MongoURL))
	if err != nil {
		logger.Error(err, "Unable to establish connection with MongoDB")
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

//nolint:govet
// CreateMonitoringUser is a method to create monitoring user inside MongoDB
func CreateMonitoringUser(params MongoDBParameters) error {
	client := InitiateMongoClient(params)
	response := client.Database(dbName).RunCommand(context.Background(), bson.D{
		{"createUser", monitoringUser}, {"pwd", params.Password},
		{"roles", []bson.M{{"role": "clusterMonitor", "db": "admin"}, {"role": "read", "db": "local"}}}},
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

//nolint:govet
// GetMongoDBUser is a method to check if user exists in MongoDB
func GetMongoDBUser(params MongoDBParameters) (bool, error) {
	client := InitiateMongoClient(params)
	collection := client.Database("admin").Collection("system.users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Count().SetMaxTime(2 * time.Second)
	docsCount, err := collection.CountDocuments(ctx, bson.D{{"user", *params.UserName}}, opts)
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
	for node := 1; node <= int(*params.ClusterNodes); node++ {
		mongoNodeInfo = append(mongoNodeInfo, bson.M{"_id": node, "host": getMongoNodeInfo(params, node)})
	}
	config := bson.M{
		"_id":     params.Name,
		"members": mongoNodeInfo,
	}
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

// getMongoNodeInfo is a method to get info for MongoDB node
func getMongoNodeInfo(params MongoDBParameters, count int) string {
	return fmt.Sprintf("%s-cluster-%v.%s-cluster.%s:27017", params.Name, count, params.Name, params.Namespace)
}

// logGenerator is a method to generate logging interface
func logGenerator(name, namespace, resourceType string) logr.Logger {
	reqLogger := log.WithValues("Namespace", namespace, "Name", name, "Resource Type", resourceType)
	return reqLogger
}
