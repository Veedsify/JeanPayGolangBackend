package utils

import (
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
}

func GetRedisClient() *redis.Client {
	return NewRedisClient()
}

func CloseRedisClient(client *redis.Client) error {
	return client.Close()
}

func PingRedisClient(client *redis.Client) error {
	return client.Ping(client.Context()).Err()
}

func SetRedisKey(client *redis.Client, key string, value interface{}, expiration time.Duration) error {
	return client.Set(client.Context(), key, value, expiration).Err()
}

func GetRedisValue(client *redis.Client, key string) (string, error) {
	return client.Get(client.Context(), key).Result()
}

func DeleteRedisKey(client *redis.Client, key string) error {
	return client.Del(client.Context(), key).Err()
}
