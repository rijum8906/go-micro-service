package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/services/organization-service/app/config"
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

func initFgaClient(ctx context.Context, config *config.Env) (*coreopenfga.Client, *apperror.AppError) {
	fgaClient, err := coreopenfga.NewClient(config.FGAAPIURL)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create FGA client").WithDetail("error", err.Error())
	}

	storeManager := coreopenfga.NewStoreManager(fgaClient.Client)
	res, err := storeManager.Create(ctx, "organziation-auth-store")
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create FGA store").WithDetail("error", err.Error())
	}
	if res == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create FGA store")
	}
	fgaClient.StoreID = res.Id

	modelManager := coreopenfga.NewModelManager(fgaClient.Client, storeManager)
	err = modelManager.Write(ctx, "organization")
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create FGA model").WithDetail("error", err.Error())
	}
	fgaClient.AuthorizationModelID = modelManager.GetAuthorizationModelID()

	return fgaClient, nil
}
