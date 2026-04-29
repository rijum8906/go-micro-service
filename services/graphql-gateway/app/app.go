// Package app contains the GraphQL Gateway application
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/resolver"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ApplicationInfra struct {
	cache *redis.Client
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

func NewApplication(ctx context.Context) (*Application, *apperror.AppError) {
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
	if appErr = app.initLogger(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initCache(ctx); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initUtils(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initGRPCClients(); appErr != nil {
		return nil, appErr
	}

	if appErr = app.initHTTPServer(); appErr != nil {
		return nil, appErr
	}

	apperror.SetConfig(apperror.Config{
		Logger: app.utils.logger,
		AppEnv: app.config.AppEnv,
		Debug:  true,
	})

	return app, nil
}

func (a *Application) initHTTPServer() *apperror.AppError {
	fmt.Println("origins:", a.config.CorsAllowedOrigins)
	fmt.Println("methods:", a.config.CorsAllowedMethods)
	fmt.Println("headers:", a.config.CorsAllowedHeaders)
	listener, err := net.Listen("tcp", net.JoinHostPort("", a.port()))
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to listen for HTTP server").WithDetail("error", err.Error())
	}

	res := resolver.NewResolver(&resolver.GrpcClients{
		AuthClient:    a.clients.AuthClient,
		SessionClient: a.clients.SessionClient,
		UserClient:    a.clients.UserClient,
		TaskClient:    a.clients.TaskClient,
	}, a.utils.token)
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: res}))

	mux := http.NewServeMux()
	mux.Handle("/query", srv)

	if a.config.AppEnv == "development" {
		mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	a.listener = listener
	a.server = &http.Server{
		Addr:    net.JoinHostPort("", a.port()),
		Handler: middleware.CORS(middleware.WithRequestHeaders(mux), a.config),
	}

	return nil
}
