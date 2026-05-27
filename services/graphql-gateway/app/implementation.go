package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
	"go.uber.org/zap"
)

func (a *Application) initUtils() *apperror.AppError {
	// Token
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(a.config.JWTSecret),
		ScopedSecret:   []byte(a.config.ScopedSecret),
		SessionTTL:     a.config.SessionTTL,
		ScopedTokenTTL: a.config.ScopedTokenTTL,
	}, a.infra.redisClient)
	a.utils.token = tokenManager

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

func (a *Application) initInfra(ctx context.Context) *apperror.AppError {
	// Redis
	cache, appErr := initCache(ctx, a.config)
	if appErr != nil {
		return appErr
	}
	a.infra.redisClient = cache

	return nil
}

func (a *Application) Shutdown(ctx context.Context) {
	if a.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
	}

	if a.clients != nil && a.clients.UserConn != nil {
		_ = a.clients.UserConn.Close()
	}

	if a.clients != nil && a.clients.TaskConn != nil {
		_ = a.clients.TaskConn.Close()
	}

	if a.infra != nil && a.infra.redisClient != nil {
		_ = a.infra.redisClient.Close()
	}
}

func (a *Application) port() string {
	return strconv.Itoa(a.config.Port)
}

func (a *Application) Addr() string {
	return fmt.Sprintf(":%s", a.port())
}

func (a *Application) UserServiceAddr() string {
	return a.config.UserServiceAddr
}

func (a *Application) GetLogger() *zap.Logger {
	return a.utils.logger
}

func (a *Application) GetConfig() *config.Env {
	return a.config
}
