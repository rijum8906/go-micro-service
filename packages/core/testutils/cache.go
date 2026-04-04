package testutils

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/cache"
)

const (
	RedisHost     = "localhost"
	RedisPort     = 6380
	RedisPassword = ""
	RedisDB       = 0
	RedisRetry    = 5
)

func MustConnectRedis() *redis.Client {
	redisClient, appErr := cache.Connect(context.Background(), cache.Config{
		Host:        RedisHost,
		Port:        RedisPort,
		Password:    RedisPassword,
		DB:          RedisDB,
		RetryCounts: RedisRetry,
	})
	if appErr != nil {
		panic(appErr)
	}

	return redisClient
}
