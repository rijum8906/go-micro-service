package app

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ApplicationInfra struct {
	cache        *redis.Client
	database     *pgxpool.Pool
	brokerClient broker.Client
}

type ApplicationUtils struct {
	logger *zap.Logger
}

type ApplicationServices struct{}

type Application struct {
	config   *config.Env
	infra    *ApplicationInfra
	utils    *ApplicationUtils
	services *ApplicationServices
	listener net.Listener
	server   *grpc.Server
}

func NewApplication(ctx context.Context) (*Application, *apperror.AppError) {
	app := &Application{
		infra:    &ApplicationInfra{},
		utils:    &ApplicationUtils{},
		services: &ApplicationServices{},
	}

	var appErr *apperror.AppError

	app.config, appErr = config.LoadEnv()
	if appErr != nil {
		return nil, appErr
	}

	// Initialize Dependencies
	if appErr = app.initInfra(ctx); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	if appErr = app.initUtils(); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	if appErr = app.initServices(); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	if appErr = app.initHandler(); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	if appErr = app.initGRPCServer(); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	apperror.SetConfig(apperror.Config{
		AppEnv: app.config.AppEnv,
		Debug:  true,
		Logger: app.utils.logger,
	})

	return app, nil
}
