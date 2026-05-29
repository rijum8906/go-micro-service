package registry

import (
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rijum8906/relay/packages/core/apperror"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/app"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/resolver"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(application *app.Application) *apperror.AppError {
	config := application.Config()

	userConn, err := grpc.NewClient(config.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return apperror.ErrThirdParty.WithMessage("failed to connect to user service").WithDetail("error", err.Error())
	}

	taskConn, err := grpc.NewClient(config.TaskServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		return apperror.ErrThirdParty.WithMessage("failed to connect to task service").WithDetail("error", err.Error())
	}

	authClient := authv1.NewAuthServiceClient(userConn)
	userClient := userv1.NewUserServiceClient(userConn)
	sessionClient := sessionv1.NewSessionServiceClient(userConn)
	taskClient := taskv1.NewTaskServiceClient(taskConn)

	listener, err := net.Listen("tcp", net.JoinHostPort("", fmt.Sprint(config.Port)))
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to listen for HTTP server").WithDetail("error", err.Error())
	}

	res := resolver.NewResolver(&resolver.GrpcClients{
		AuthClient:    authClient,
		SessionClient: sessionClient,
		UserClient:    userClient,
		TaskClient:    taskClient,
	}, application.TokenManager())
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: res}))

	mux := http.NewServeMux()
	mux.Handle("/query", srv)

	if config.AppEnv == "development" {
		mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	server := &http.Server{
		Addr:    net.JoinHostPort("", fmt.Sprint(config.Port)),
		Handler: middleware.CORS(middleware.WithRequestHeaders(mux), config),
	}

	if err = server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return apperror.ErrInternal.WithMessage("failed to start HTTP server").WithDetail("error", err.Error())
	}

	return nil
}
