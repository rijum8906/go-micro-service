package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/notification-service/app"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	application, appErr := app.NewApplication(ctx)
	logger := application.GetLogger()

	if appErr != nil {
		// If logger is nil, fallback to standard log
		if logger == nil {
			log.Fatalf("failed to create application: %v", appErr)
		}
		logger.Fatal("failed to create application", zap.Error(appErr))
	}

	runErrCh := make(chan *apperror.AppError, 1)
	go func() {
		runErrCh <- application.Run()
	}()

	logger.Info("service started",
		zap.String("service", application.GetConfig().AppName),
		zap.String("env", application.GetConfig().AppEnv),
		zap.Int("port", application.GetConfig().Port),
	)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		logger.Info("shutdown signal received",
			zap.String("signal", sig.String()),
		)

		application.Shutdown()

		if appErr = <-runErrCh; appErr != nil {
			logger.Error("shutdown completed with server error",
				zap.Error(appErr),
			)
			os.Exit(1)
		}

		logger.Info("shutdown complete")

	case appErr = <-runErrCh:
		if appErr != nil {
			logger.Error("service stopped with error",
				zap.Error(appErr),
			)
			os.Exit(1)
		}

		logger.Info("service stopped gracefully")
	}
}
