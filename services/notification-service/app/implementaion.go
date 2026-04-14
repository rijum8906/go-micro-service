// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/nats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (a *Application) initDB(ctx context.Context) *apperror.AppError {
	pool, appErr := db.Connect(ctx, db.Config{
		Host:        a.config.DBHost,
		Port:        a.config.DBPort,
		User:        a.config.DBUser,
		Password:    a.config.DBPassword,
		DBName:      a.config.DBName,
		SSLMode:     a.config.DBSSLMode,
		RetryCounts: 5,
	})
	if appErr != nil {
		return appErr
	}

	a.infra.database = pool
	return nil
}

func (a *Application) initCache(ctx context.Context) *apperror.AppError {
	cache, appErr := cache.Connect(ctx, cache.Config{
		Host:        a.config.RedisHost,
		Port:        a.config.RedisPort,
		DB:          0,
		Password:    a.config.RedisPass,
		RetryCounts: 5,
	})

	if appErr != nil {
		return appErr
	}

	a.infra.cache = cache
	return nil
}

func (a *Application) initNATS(ctx context.Context) *apperror.AppError {
	client, appErr := nats.Connect(ctx, nats.Config{
		URL:        a.config.NATSURL,
		ClientName: a.config.NATSClientName,
	})
	if appErr != nil {
		return appErr
	}

	a.infra.nats = client
	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	return nil
}

func (a *Application) initHandler() *apperror.AppError {
	return nil
}

func (a *Application) initGRPCServer() *apperror.AppError {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Port))
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to listen for gRPC server").WithDetail("error", err.Error())
	}
	a.listener = listener

	// Create and Register grpc server
	server := grpc.NewServer()
	a.server = server

	if a.config.AppEnv == "development" {
		reflection.Register(server)
	}

	return nil
}

func (a *Application) Run() *apperror.AppError {
	err := a.server.Serve(a.listener)
	if err != nil {
		if err == grpc.ErrServerStopped {
			return nil
		}

		return apperror.ErrInternal.WithMessage("failed to start gRPC server").WithDetail("error", err.Error())
	}

	return nil
}

func (a *Application) Shutdown() {
	if a.server != nil {
		a.server.GracefulStop()
	}

	if a.infra != nil && a.infra.database != nil {
		a.infra.database.Close()
	}

	if a.infra != nil && a.infra.cache != nil {
		a.infra.cache.Close()
	}

	if a.infra != nil && a.infra.nats != nil {
		_ = a.infra.nats.Drain()
	}
}
