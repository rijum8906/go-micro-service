// Package app contains the GraphQL Gateway application
package app

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	instance *Application
	once     sync.Once
	initErr  *apperror.AppError
)

type ApplicationInfra struct {
	redisClient *redis.Client
}

type ApplicationUtils struct {
	token  *token.TokenManager
	logger *zap.Logger
}

type GrpcClients struct {
	UserConn      *grpc.ClientConn
	TaskConn      *grpc.ClientConn
	AuthClient    authv1.AuthServiceClient
	UserClient    userv1.UserServiceClient
	SessionClient sessionv1.SessionServiceClient
	TaskClient    taskv1.TaskServiceClient
}

type Application struct {
	config   *config.Env
	infra    *ApplicationInfra
	utils    *ApplicationUtils
	clients  *GrpcClients
	listener net.Listener
	server   *http.Server
}

func GetInstance() (*Application, *apperror.AppError) {
	once.Do(func() {
		ctx := context.Background()
		instance, initErr = newApplication(ctx)
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func newApplication(ctx context.Context) (*Application, *apperror.AppError) {
	app := &Application{
		infra: &ApplicationInfra{},
		utils: &ApplicationUtils{},
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

	apperror.SetConfig(&apperror.Config{
		Logger: app.utils.logger,
		AppEnv: app.config.AppEnv,
		Debug:  true,
	})

	return app, nil
}
