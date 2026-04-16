// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/services/notification-service/internal/handler/broker"
	"github.com/rijum8906/relay/services/notification-service/internal/services/subscriber"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NOTE: do not use the logger here

var mailerConfig mailer.Config

func (a *Application) initInfra(ctx context.Context) *apperror.AppError {
	// PostgreSQL
	database, appErr := initDB(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.database = database

	// Redis
	cache, appErr := initCache(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.cache = cache

	// NATS
	nats, appErr := initNATS(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.nats = nats

	// Mailer
	mailer, appErr := initMailer(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	mailerConfig = getMailerConfig(a.config)
	a.infra.mailer = mailer

	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	logger, appErr := initLogger(a.config)
	if appErr != nil {
		return appErr
	}
	a.utils.logger = logger

	tm, appErr := initTemplateManager(a.config)
	if appErr != nil {
		return appErr
	}
	a.utils.tm = tm

	return nil
}

func (a *Application) initServices() *apperror.AppError {
	if mailerConfig == (mailer.Config{}) {
		return apperror.ErrInternal.WithMessage("app infra is not initialized")
	}

	subsciberService, appErr := subscriber.New(a.infra.nats, "verification", a.utils.logger, mailerConfig)
	if appErr != nil {
		return appErr
	}
	a.services.subscriberService = subsciberService

	return nil
}

func (a *Application) initHandler() *apperror.AppError {
	subscriberHandler, appErr := broker.New(a.services.subscriberService, a.infra.nats, &mailerConfig)
	if appErr != nil {
		return appErr
	}

	go func(handler *broker.SubscribeHandler) {
		handler.Subscribe()
	}(subscriberHandler)

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

func (a *Application) GetLogger() *zap.Logger {
	return a.utils.logger
}

func (a *Application) GetConfig() *env.Config {
	return a.config
}
