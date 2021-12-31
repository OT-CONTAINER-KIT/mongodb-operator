package mongogo

import (
	"context"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_mongo")

const (
	dbName         = "admin"
	monitoringUser = "monitoring"
)

// MongoDBParameters is a struct for MongoDB related inputs
type MongoDBParameters struct {
	MongoURL  string
	Namespace string
	Name      string
	Password  string
}

// InitiateMongoClient is a method to create client connection with MongoDB
func InitiateMongoClient(params MongoDBParameters) mongo.Client {
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
func DiscconnectMongoClient(client mongo.Client) error {
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			return err
		}
	}()
	return nil
}

// CreateMonitoringUser is a method to create monitoring user inside MongoDB
func CreateMonitoringUser(params MongoDBParameters) error {
	client := InitiateMongoClient(params)
	response := client.Database(dbName).RunCommand(context.Background(), bson.D{
		{"createUser", monitoringUser}, {"pwd", password},
		{"roles", []bson.M{{"role": "clusterMonitor", "db": "admin"}, {"role": "read", "db": "local"}}}},
	)
	if response.Err() != nil {
		logger.Error(err, "Unable to create user for MongoDB")
		return err
	}
	err = DiscconnectMongoClient(client)
	if response.Err() != nil {
		logger.Error(err, "Unable to disconnect from MongoDB")
		return err
	}
	return nil
}

// logGenerator is a method to generate logging interface
func logGenerator(name, namespace, resourceType string) logr.Logger {
	reqLogger := log.WithValues("Namespace", namespace, "Name", name, "Resource Type", resourceType)
	return reqLogger
}
