package app

import (
	"context"
	"fmt"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/user/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

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

	return nil
}

func (a *Application) initUtils() *apperror.AppError {
	if a.infra.cache == nil {
		return apperror.ErrInternal.WithMessage("failed to initialize token manager").WithDetail("error", "redis client is nil")
	}

	// Token
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(a.config.JWTSecret),
		ScopedSecret:   []byte(a.config.ScopedSecret),
		SessionTTL:     a.config.SessionTTL,
		ScopedTokenTTL: a.config.ScopedTokenTTL,
	}, a.infra.cache)
	a.utils.token = tokenManager

	// Hash
	hashService := hash.NewHashService(hash.Config{
		BcryptCost: 10,
	})
	a.utils.hash = hashService

	// Logger
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

func (a *Application) GetLogger() *zap.Logger {
	return a.utils.logger
}

func (a *Application) GetConfig() *config.Env {
	return a.config
}
