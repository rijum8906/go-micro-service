package handlers

import (
	"buf.build/go/protovalidate"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	"github.com/rijum8906/relay/services/user-service/internal/services/account"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/services/profile"
)

type Services struct {
	AuthService    auth.AuthService
	AccountService account.AccountService
	Profileservice profile.ProfileService
}

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	authService       auth.AuthService
	accountService    account.AccountService
	middlewareService middleware.Middleware
	validator         protovalidate.Validator
}

func NewAuthHandler(services *Services, middlewareService middleware.Middleware) *AuthHandler {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}

	return &AuthHandler{
		authService:       services.AuthService,
		accountService:    services.AccountService,
		middlewareService: middlewareService,
		validator:         validator,
	}
}
