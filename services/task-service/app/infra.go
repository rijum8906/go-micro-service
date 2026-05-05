package app

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/services/task-service/app/config"
)

func initDB(ctx context.Context, config *config.Env) (*pgxpool.Pool, *apperror.AppError) {
	pool, appErr := db.Connect(ctx, db.Config{
		Host:        config.DBHost,
		Port:        config.DBPort,
		User:        config.DBUser,
		Password:    config.DBPassword,
		DBName:      config.DBName,
		SSLMode:     config.DBSSLMode,
		RetryCounts: 5,
	})
	if appErr != nil {
		return nil, appErr
	}

	return pool, nil
}

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

func initNATSClient(config *config.Env) (broker.Client, *apperror.AppError) {
	client := broker.NewClient()
	client.Connect(config.NATSURL)

	return client, nil
}

func initOpenFGA(config *config.Env) (*coreopenfga.Client, coreopenfga.TuppleManager, *apperror.AppError) {
	if strings.TrimSpace(config.OpenFGAURL) == "" {
		return nil, nil, nil
	}

	fgaClient, appErr := coreopenfga.NewClient(config.OpenFGAURL)
	if appErr != nil {
		return nil, nil, appErr
	}

	fgaClient.StoreID = config.OpenFGAStoreID
	fgaClient.AuthorizationModelID = config.OpenFGAAuthModelID

	return fgaClient, coreopenfga.NewTupleManager(fgaClient), nil
}
