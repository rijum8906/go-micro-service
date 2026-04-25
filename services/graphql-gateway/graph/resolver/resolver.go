package resolver

import (
	"github.com/go-playground/validator/v10"
	"github.com/rijum8906/relay/packages/core/token"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type GrpcClients struct {
	AuthClient    authv1.AuthServiceClient
	UserClient    userv1.UserServiceClient
	SessionClient sessionv1.SessionServiceClient
	TaskClient    taskv1.TaskServiceClient
}

type Resolver struct {
	Clients  *GrpcClients
	Validate *validator.Validate
	Token    *token.TokenManager
}

func NewResolver(clients *GrpcClients, token *token.TokenManager) *Resolver {
	return &Resolver{
		Validate: validator.New(),
		Clients:  clients,
		Token:    token,
	}
}
