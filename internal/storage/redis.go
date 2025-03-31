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
