package app

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/nats"
	"google.golang.org/grpc"
)

type ApplicationInfra struct {
	cache    *redis.Client
	database *pgxpool.Pool
	nats     *nats.Client
}

type ApplicationUtils struct{}

type ApplicationServices struct{}

type Application struct {
	config   *env.Config
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

	app.config, appErr = env.Load()
	if appErr != nil {
		return nil, appErr
	}

	apperror.SetConfig(apperror.Config{
		AppEnv: app.config.AppEnv,
	})

	// Initialize Dependencies
	if appErr = app.initDB(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initCache(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initNATS(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initUtils(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initHandler(); appErr != nil {
		fmt.Println(appErr.Details)
		return nil, appErr
	}

	if appErr = app.initGRPCServer(); appErr != nil {
		return nil, appErr
	}

	return app, nil
}
