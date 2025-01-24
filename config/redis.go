package config

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: GetRedisAddress(),
	})

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}
