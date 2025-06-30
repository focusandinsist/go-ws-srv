package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	dbName string
	client *mongo.Client
	// collection *mongo.Collection
}

// 新建个表
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

	// collection := client.Database(dbName).Collection(collectionName)
	return &MongoStorage{client: client, dbName: dbName}, nil
}

// 获取集合
func (ms *MongoStorage) Collection(name string) *mongo.Collection {
	return ms.client.Database(ms.dbName).Collection(name)
}

// 存储数据
func (ms *MongoStorage) StoreMessage(collection string, msg any) error {
	_, err := ms.Collection(collection).InsertOne(context.Background(), msg)
	return err
}

// GetMessages 根据条件获取消息
func (ms *MongoStorage) GetMessages(collection string, filter any) ([]any, error) {
	cursor, err := ms.Collection(collection).Find(context.Background(), filter)
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
