package app

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
)

func initCache(ctx context.Context, config *config.Env) (*redis.Client, *apperror.AppError) {
	cache, appErr := cache.Connect(ctx, cache.Config{
		Host:        config.RedisHost,
		Port:        config.RedisPort,
		DB:          config.RedisDB,
		Password:    config.RedisPass,
		RetryCounts: 5,
	})

	if appErr != nil {
		return nil, appErr
	}

	return cache, nil
}
