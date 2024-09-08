package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func InitDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbUri := os.Getenv("DB_URI")

	if dbUri == "" {
		log.Fatal("Database URL is empty")
	}

	clientOpts := options.Client().ApplyURI(dbUri)

	var err error
	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %v", err)
	}

	// Check if connection is successful
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging mongodb: %v", err)
	}

	log.Println("Connected to database")

	// NOTE: UNCOMMENT THIS TO CREATE SOME UNIQUE INDEX
	// var collection *mongo.Collection
	//
	// collection, err = GetCollection("guild_settings")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// setupDatabaseIndexes(ctx, collection)
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	if client == nil {
		return nil, fmt.Errorf("database client not initialized")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, fmt.Errorf("missing DB_NAME in environment")
	}

	collection := client.Database(dbName).Collection(collectionName)
	return collection, nil
}

func CloseDatabase() {
	if client == nil {
		log.Println("Database client is already nil")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("Error stopping mongodb: %v", err)
	}

	log.Println("Disconnected from mongodb")
	client = nil
}

func setupDatabaseIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"guild_id": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("error creating unique index: %v", err)
	}

	return nil
}
