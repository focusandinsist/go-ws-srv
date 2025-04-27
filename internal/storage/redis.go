package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage{client: client}
}

func (rs *RedisStorage) Set(key string, value interface{}) error {
	return rs.client.Set(context.Background(), key, value, 0).Err()
}

func (rs *RedisStorage) Get(key string) (string, error) {
	return rs.client.Get(context.Background(), key).Result()
}

func (rs *RedisStorage) AddOfflineMessage(userID string, message string) error {
	return rs.client.RPush(context.Background(), "offline:"+userID, message).Err()
}

func (rs *RedisStorage) GetOfflineMessages(userID string) ([]string, error) {
	return rs.client.LRange(context.Background(), "offline:"+userID, 0, -1).Result()
}

func (rs *RedisStorage) ClearOfflineMessages(userID string) error {
	return rs.client.Del(context.Background(), "offline:"+userID).Err()
}
