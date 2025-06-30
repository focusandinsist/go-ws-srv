package storage

import (
	"context"
	"encoding/json"
	"strconv"
	"websocket-server/protocol"

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

func (rs *RedisStorage) AddOfflineMessage(userID int64, message *protocol.Message) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return rs.client.RPush(context.Background(), "offline:"+strconv.FormatInt(userID, 10), b).Err()
}

func (rs *RedisStorage) GetOfflineMessages(userID int64) ([]string, error) {
	return rs.client.LRange(context.Background(), "offline:"+strconv.FormatInt(userID, 10), 0, -1).Result()
}

func (rs *RedisStorage) ClearOfflineMessages(userID int64) error {
	return rs.client.Del(context.Background(), "offline:"+strconv.FormatInt(userID, 10)).Err()
}
