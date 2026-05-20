package app

import (
	"context"
	"net"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/app/config"
	"github.com/rijum8906/relay/services/notification-service/internal/services/subscriber"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	instance *Application
	once     sync.Once
	appErr   *apperror.AppError
)

type ApplicationInfra struct {
	cache        *redis.Client
	database     *pgxpool.Pool
	brokerClient broker.Client
	mailer       *mail.Client
}

type ApplicationUtils struct {
	logger *zap.Logger
	tm     template.TemplateManager
}

type ApplicationServices struct {
	subscriberService subscriber.Service
}

type ApplicationState struct {
	isLogggerLoaded bool
}

type Application struct {
	state    *ApplicationState
	config   *config.Env
	infra    *ApplicationInfra
	utils    *ApplicationUtils
	services *ApplicationServices
	listener net.Listener
	server   *grpc.Server
}

func GetInstance() (*Application, *apperror.AppError) {
	once.Do(func() {
		instance, appErr = newApplication(context.Background())
	})
	if appErr != nil {
		return nil, appErr
	}

	return instance, nil
}

func newApplication(ctx context.Context) (*Application, *apperror.AppError) {
	app := &Application{
		infra:    &ApplicationInfra{},
		utils:    &ApplicationUtils{},
		services: &ApplicationServices{},
		state: &ApplicationState{
			isLogggerLoaded: false,
		},
	}

	var appErr *apperror.AppError

	app.config, appErr = config.LoadEnv()
	if appErr != nil {
		return nil, appErr
	}

	// Initialize Dependencies

	if appErr = app.initUtils(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initInfra(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initServices(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initHandler(); appErr != nil {
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
