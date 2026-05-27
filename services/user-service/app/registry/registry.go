// Package registry
package registry

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"google.golang.org/grpc/reflection"
)

func Run(application *app.Application) *apperror.AppError {
	server := application.GRPCServer()

	// User Service
	userService, appErr := user.New()
	if appErr != nil {
		return appErr
	}
	userv1.RegisterUserServiceServer(server, userService)

	// Auth Service
	authService, appErr := auth.New()
	if appErr != nil {
		return appErr
	}
	authv1.RegisterAuthServiceServer(server, authService)

	// Session Service
	sessionService, appErr := session.New()
	if appErr != nil {
		return appErr
	}
	sessionv1.RegisterSessionServiceServer(server, sessionService)

	// Enable reflection
	if application.Config().AppEnv == "development" {
		reflection.Register(server)
	}

	return application.Run()
}
