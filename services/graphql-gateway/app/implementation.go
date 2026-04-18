package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *Application) initLogger() *apperror.AppError {
	var zapConfig zap.Config

	if a.config.AppEnv == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	if a.config.EnableJSON {
		zapConfig.Encoding = "json"
	}

	if a.config.LogLevel != "" {
		level, err := zapcore.ParseLevel(a.config.LogLevel)
		if err != nil {
			level = zapcore.InfoLevel
		}
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	}

	zapConfig.DisableCaller = !a.config.EnableCaller
	zapConfig.DisableStacktrace = !a.config.EnableStack

	// Configure output paths
	if a.config.LogFile != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(a.config.LogFile), 0o755); err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to create log directory").
				WithDetail("error", err.Error())
		}

		zapConfig.OutputPaths = []string{"stdout", a.config.LogFile}
		zapConfig.ErrorOutputPaths = []string{"stderr", a.config.LogFile}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	// Build the logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to create zap logger").
			WithDetail("error", err.Error())
	}

	a.utils.logger = logger

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

func (a *Application) initUtils() *apperror.AppError {
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(a.config.JWTSecret),
		ScopedSecret:   []byte(a.config.ScopedSecret),
		SessionTTL:     a.config.SessionTTL,
		ScopedTokenTTL: a.config.ScopedTokenTTL,
	}, a.infra.cache)
	a.utils.token = tokenManager

	return nil
}

func (a *Application) initGRPCClients() *apperror.AppError {
	conn, err := grpc.NewClient(a.config.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return apperror.ErrThirdParty.WithMessage("failed to connect to user service").WithDetail("error", err.Error())
	}

	// Initialize gRPC clients
	authClient := authv1.NewAuthServiceClient(conn)
	userClinet := userv1.NewUserServiceClient(conn)
	sessionClient := sessionv1.NewSessionServiceClient(conn)
	a.clients = &GrpcClients{
		Conn:          conn,
		AuthClient:    authClient,
		UserClient:    userClinet,
		SessionClient: sessionClient,
	}

	return nil
}

func (a *Application) Run() *apperror.AppError {
	err := a.server.Serve(a.listener)
	if err != nil && err != http.ErrServerClosed {
		return apperror.ErrInternal.WithMessage("failed to start HTTP server").WithDetail("error", err.Error())
	}

	return nil
}

func (a *Application) Shutdown(ctx context.Context) {
	if a.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
	}

	if a.clients != nil && a.clients.Conn != nil {
		_ = a.clients.Conn.Close()
	}

	if a.infra != nil && a.infra.cache != nil {
		_ = a.infra.cache.Close()
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
