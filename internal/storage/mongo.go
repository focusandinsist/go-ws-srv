package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoStorage(uri, dbName, collectionName string) (*MongoStorage, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &MongoStorage{client: client, collection: collection}, nil
}

func (ms *MongoStorage) StoreMessage(msg interface{}) error {
	_, err := ms.collection.InsertOne(context.Background(), msg)
	return err
}

func (ms *MongoStorage) GetMessages(filter interface{}) ([]interface{}, error) {
	cursor, err := ms.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []interface{}
	for cursor.Next(context.Background()) {
		var msg interface{}
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
