package app

import (
	"context"
	"net"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ApplicationUtils struct {
	logger *zap.Logger
}

type ApplicationServices struct {
}

type Application struct {
	config   *config.Config
	utils    *ApplicationUtils
	services *ApplicationServices
	listener net.Listener
	server   *grpc.Server
}

func NewApplication(ctx context.Context) (*Application, *apperror.AppError) {
	_ = ctx

	app := &Application{
		utils:    &ApplicationUtils{},
		services: &ApplicationServices{},
	}

	var appErr *apperror.AppError

	app.config, appErr = config.Load()
	if appErr != nil {
		return nil, appErr
	}

	// Initialize Dependencies
	if appErr = app.initLogger(); appErr != nil {
		return nil, appErr
	}
	if appErr = app.initGRPCServer(); appErr != nil {
		return nil, appErr
	}

	apperror.SetConfig(apperror.Config{
		AppEnv: app.config.AppEnv,
		Debug:  true,
		Logger: app.utils.logger,
	})

	return app, nil
}
