// Package handlers contains the gRPC handlers for the auth service
package handlers

import (
	"context"
	"errors"

	// Import the generated code from your shared packages folder
	"buf.build/go/protovalidate"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/grpc/metadata"
)

type AuthHandler struct {
	// Embedding this is a gRPC requirement for forward compatibility
	user_servicev1.UnimplementedAuthServiceServer
	authService       auth.AuthService
	middlewareService middleware.Middleware
	validator         protovalidate.Validator
}

func NewAuthHandler(authService auth.AuthService, middlewareService middleware.Middleware) *AuthHandler {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}

	return &AuthHandler{
		authService:       authService,
		middlewareService: middlewareService,
		validator:         validator,
	}
}

func (h *AuthHandler) Signin(ctx context.Context, req *user_servicev1.SigninRequest) (*user_servicev1.AuthResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, errors.New("validation error")
	}
	result, err := h.authService.Signin(ctx, req)
	if err != nil {
		return nil, errors.New(err.Message)
	}

	return result, nil
}

func (h *AuthHandler) Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, errors.New("validation error")
	}
	result, err := h.authService.SignUp(ctx, req)
	if err != nil {
		return nil, errors.New(err.Message)
	}

	return result, nil
}

func (h *AuthHandler) Signout(ctx context.Context, req *user_servicev1.SignoutRequest) (*user_servicev1.SignoutResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, errors.New("validation error")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	authz := md.Get("x-user-id")
	userID, appErr := utils.StrIDToPgUUID(authz[0])
	if !userID.Valid || appErr != nil {
		return nil, errors.New("missing authorization header")
	}

	authzMeta := request.AuthzMetadata{
		UserID: userID,
	}

	appErr = h.authService.Signout(ctx, req, authzMeta)
	if appErr != nil {
		return nil, errors.New(appErr.Message)
	}

	return &user_servicev1.SignoutResponse{
		Success: true,
	}, nil
}
