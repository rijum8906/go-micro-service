package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/packages/core/apperror"
	"go.uber.org/zap"
)

// RunService starts the task service and blocks until it exits or receives a shutdown signal.
func RunService(ctx context.Context) error {
	application, appErr := NewApplication(ctx)
	if appErr != nil {
		if application != nil {
			if logger := application.GetLogger(); logger != nil {
				logger.Error("failed to create application", zap.Error(appErr))
			}
		}

		log.Printf("failed to create application: %v", appErr)
		return fmt.Errorf("create application: %w", appErr)
	}

	logger := application.GetLogger()
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
			return fmt.Errorf("shutdown completed with server error: %w", appErr)
		}

		logger.Info("shutdown complete")
	case appErr = <-runErrCh:
		if appErr != nil {
			logger.Error("service stopped with error",
				zap.Error(appErr),
			)
			return fmt.Errorf("service stopped with error: %w", appErr)
		}

		logger.Info("service stopped gracefully")
	}

	return nil
}
