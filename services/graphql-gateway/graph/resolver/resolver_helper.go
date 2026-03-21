package resolver

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/common/database/redis"
	"github.com/rijum8906/relay/packages/common/env"
)

func connectRedis(ctx context.Context, appEnv *env.Env) (*goredis.Client, error) {
	redisClient, appErr := redis.Connect(ctx, redis.Config{
		Database: appEnv.RedisDatabase,
		Host:     appEnv.RedisHost,
		Port:     appEnv.RedisPort,
		User:     appEnv.RedisUser,
		Password: appEnv.RedisPassword,
	})
	if appErr != nil {
		return nil, fmt.Errorf("connect redis: %w", appErr)
	}

	return redisClient, nil
}
