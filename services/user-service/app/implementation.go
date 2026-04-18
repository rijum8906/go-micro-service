package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app/config"
	userdb "github.com/rijum8906/relay/services/user/internal/db"
	handler "github.com/rijum8906/relay/services/user/internal/handlers/grpc"
	profilerepo "github.com/rijum8906/relay/services/user/internal/repository/profile"
	sessionrepo "github.com/rijum8906/relay/services/user/internal/repository/session"
	userrepo "github.com/rijum8906/relay/services/user/internal/repository/user"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	client, appErr := broker.Connect(ctx, broker.Config{
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
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(a.config.JWTSecret),
		ScopedSecret:   []byte(a.config.ScopedSecret),
		SessionTTL:     a.config.SessionTTL,
		ScopedTokenTTL: a.config.ScopedTokenTTL,
	}, a.infra.cache)
	a.utils.token = tokenManager

	hashService := hash.NewHashService(hash.Config{
		BcryptCost: 10,
	})
	a.utils.hash = hashService

	return nil
}

func (a *Application) initHandler() *apperror.AppError {
	queries := userdb.New(a.infra.database)
	repos := &utils.Repos{
		User:    userrepo.NewAuthRepository(queries),
		Profile: profilerepo.NewProfileRepository(queries),
		Session: sessionrepo.NewSessionRepository(queries),
	}

	authService, appErr := auth.NewAuthService(repos, utils.NewUtils(a.utils.token, a.utils.hash), &a.config.CoreEnv, a.infra.nats)
	if appErr != nil {
		return appErr
	}

	userService, appErr := user.NewUserService(repos, utils.NewUtils(a.utils.token, a.utils.hash), &a.config.CoreEnv)
	if appErr != nil {
		return appErr
	}

	sessionService, appErr := session.NewSessionService(repos, utils.NewUtils(a.utils.token, a.utils.hash), &a.config.CoreEnv)
	if appErr != nil {
		return appErr
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	a.services.auth = authHandler
	a.services.session = sessionHandler
	a.services.user = userHandler

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
	authv1.RegisterAuthServiceServer(server, a.services.auth)
	userv1.RegisterUserServiceServer(server, a.services.user)
	sessionv1.RegisterSessionServiceServer(server, a.services.session)

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

func (a *Application) GetConfig() *config.Env {
	return a.config
}
