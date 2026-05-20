package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/services/notification-service/app"
	"go.uber.org/zap"
)

func main() {
	application, appErr := app.GetInstance()
	var logger *zap.Logger

	if appErr != nil {
		logger = corelogger.NewDevLogger()
		// If logger is nil, create a default development logger
		logger.Error(
			"failed to initialize application",
			zap.String("error", appErr.Message),
			zap.Any("details", appErr.Details),
		)
	}

	logger = application.Logger()

	runErrCh := make(chan *apperror.AppError, 1)
	go func() {
		runErrCh <- application.Run()
	}()

	logger.Info(
		"service started",
		zap.String("service", application.Config().AppName),
		zap.String("env", application.Config().AppEnv),
		zap.Int("port", application.Config().Port),
	)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		logger.Info(
			"shutdown signal received",
			zap.String("signal", sig.String()),
		)

		application.Shutdown()

		if appErr = <-runErrCh; appErr != nil {
			logger.Error(
				"shutdown completed with server error",
				zap.Error(appErr),
			)
			os.Exit(1)
		}

		logger.Info("shutdown complete")

	case appErr = <-runErrCh:
		if appErr != nil {
			logger.Error(
				"service stopped with error",
				zap.Error(appErr),
			)
			os.Exit(1)
		}

		logger.Info("service stopped gracefully")
	}
}
