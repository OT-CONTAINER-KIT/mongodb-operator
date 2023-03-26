package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createMongoDBClient() (*mongo.Client, error) {
	// Get the MongoDB connection details from the operator config
	mongoURI := "mongodb://localhost:27017"

	// i need
	//user -> adminuser and password
	// Create a new MongoDB client
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}
