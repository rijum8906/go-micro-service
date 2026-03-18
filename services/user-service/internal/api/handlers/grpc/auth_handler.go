// Package handlers contains the gRPC handlers for the auth service
package handlers

import (
	"context"
	"errors"

	// Import the generated code from your shared packages folder
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
)

type AuthHandler struct {
	// Embedding this is a gRPC requirement for forward compatibility
	user_servicev1.UnimplementedAuthServiceServer
	authService       auth.AuthService
	middlewareService middleware.Middleware
}

func NewAuthHandler(authService auth.AuthService, middlewareService middleware.Middleware) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		middlewareService: middlewareService,
	}
}

func (h *AuthHandler) Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, error) {
	result, err := h.authService.SignUp(ctx, req)
	if err != nil {
		return nil, errors.New(err.Message)
	}

	return result, nil
}
