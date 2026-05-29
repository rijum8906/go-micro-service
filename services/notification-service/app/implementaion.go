// Package app
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NOTE: do not use the logger here

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
	a.infra.mailer = mailer
	a.utils.mailerConfig = getMailerConfig(a.config)

	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	logger, appErr := corelogger.InitLogger(corelogger.LoggerConfig{
		AppEnv:       a.config.AppEnv,
		LogLevel:     a.config.LogLevel,
		LogFile:      a.config.LogFile,
		EnableJSON:   a.config.EnableJSON,
		EnableCaller: a.config.EnableCaller,
		EnableStack:  a.config.EnableStack,
	})
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
