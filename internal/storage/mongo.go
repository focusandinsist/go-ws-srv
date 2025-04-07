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
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// 确保连接成功
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查数据库连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &MongoStorage{client: client, collection: collection}, nil
}

func (ms *MongoStorage) StoreMessage(msg any) error {
	_, err := ms.collection.InsertOne(context.Background(), msg)
	return err
}

func (ms *MongoStorage) GetMessages(filter any) ([]any, error) {
	cursor, err := ms.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []any
	for cursor.Next(context.Background()) {
		var msg any
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
