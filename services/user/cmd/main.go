package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/app"
)

func main() {
	logger := log.New(os.Stdout, "user-service ", log.LstdFlags|log.LUTC|log.Lmsgprefix)
	ctx := context.Background()

	logger.Printf("starting")

	application, appErr := app.NewApplication(ctx)
	if appErr != nil {
		logger.Printf("startup failed: %v", appErr)
		os.Exit(1)
	}

	runErrCh := make(chan *apperror.AppError, 1)
	go func() {
		runErrCh <- application.Run()
	}()

	logger.Printf("started")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		logger.Printf("shutdown signal received: %s", sig)
		application.Shutdown()

		if appErr = <-runErrCh; appErr != nil {
			logger.Printf("shutdown completed with server error: %v", appErr)
			os.Exit(1)
		}

		logger.Printf("shutdown complete")
	case appErr = <-runErrCh:
		if appErr != nil {
			logger.Printf("service stopped with error: %v", appErr)
			os.Exit(1)
		}

		logger.Printf("service stopped")
	}
}
