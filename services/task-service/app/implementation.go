package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/app/config"
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
}

func (a *Application) GetLogger() *zap.Logger {
	return a.utils.logger
}

func (a *Application) GetConfig() *config.Config {
	return a.config
}
