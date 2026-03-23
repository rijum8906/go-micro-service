package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/token"
)

func bootstrap() (*env.Config, *apperror.AppError) {
	env, appErr := env.Load()
	if appErr != nil {
		return nil, appErr
	}
	return env, nil
}

func initTokenManager(env *env.Config, redisClient *redis.Client) *token.TokenManager {
	return token.NewTokenManager(env.JWTSecret, env.ScopedSecret, redisClient)
}

func initPostgres(env *env.Config, ctx context.Context) (*pgxpool.Pool, *apperror.AppError) {
	return db.Connect(ctx, db.Config{
		Host:        env.DBHost,
		Port:        env.DBPort,
		User:        env.DBUser,
		Password:    env.DBPassword,
		DBName:      env.DBName,
		SSLMode:     env.DBSSLMode,
		RetryCounts: 5,
	})
}

func initRedis(env *env.Config, ctx context.Context) (*redis.Client, *apperror.AppError) {
	return cache.Connect(ctx, cache.Config{
		Host:        env.RedisHost,
		Port:        env.RedisPort,
		DB:          0,
		Password:    "",
		RetryCounts: 5,
	})
}

func main() {
	ctx := context.Background()

	env, appErr := bootstrap()
	if appErr != nil {
		panic(appErr)
	}

	_, appErr = initPostgres(env, ctx)
	if appErr != nil {
		panic(appErr)
	}

	redisClient, appErr := initRedis(env, ctx)
	if appErr != nil {
		panic(appErr)
	}

	_ = initTokenManager(env, redisClient)
}
