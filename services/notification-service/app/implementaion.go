// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/services/notification-service/internal/handler/handler"
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
	nats, appErr := initNATSClient(a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.brokerClient = nats

	// Mailer
	mailer, appErr := initMailer(a.config)
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

	tm, appErr := initTemplateManager()
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

	brokerSubscriber := broker.NewSubscriber(a.infra.brokerClient.GetClient())

	subsciberService, appErr := subscriber.New(brokerSubscriber, a.utils.logger, mailerConfig, a.utils.tm)
	if appErr != nil {
		return appErr
	}

	a.services.subscriberService = subsciberService

	return nil
}

func (a *Application) initHandler() *apperror.AppError {
	subscriberHandler, appErr := handler.New(a.services.subscriberService, a.infra.brokerClient, &mailerConfig, a.utils.tm)
	if appErr != nil {
		return appErr
	}

	if appErr = subscriberHandler.CreateStreams(); appErr != nil {
		return appErr
	}
	if appErr = subscriberHandler.CreateConsumers(); appErr != nil {
		return appErr
	}

	go func(handler *handler.SubscribeHandler, logger *zap.Logger) {
		if appErr := handler.Subscribe(); appErr != nil {
			logger.Error("failed to subscribe to nats", zap.Error(appErr), zap.Any("detals", appErr.Details))
		}
	}(subscriberHandler, a.utils.logger)

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

	if a.infra != nil && a.infra.brokerClient != nil {
		_ = a.infra.brokerClient.Drain()
	}
}
