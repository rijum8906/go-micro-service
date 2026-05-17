package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/organization-service/app"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	application, appErr := app.NewApplication(ctx)
	if appErr != nil {
		// If logger is nil, fallback to standard log
		log.Fatalf("failed to create application: %v\nDetails:%v", appErr, appErr.Details)
	}

	logger := application.Logger()

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

		application.Cleanup()

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
