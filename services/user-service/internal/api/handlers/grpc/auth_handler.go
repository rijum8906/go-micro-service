// Package handlers contains the gRPC handlers for the auth service
package handlers

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
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
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	result, err := h.authService.Signin(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Message)
	}

	return result, nil
}

func (h *AuthHandler) Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, errors.New("validation error")
	}

	result, err := h.authService.SignUp(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Message)
	}

	return result, nil
}

func (h *AuthHandler) Signout(ctx context.Context, req *user_servicev1.SignoutRequest) (*user_servicev1.SignoutResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authz := md.Get("x-user-id")
	if len(authz) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing x-user-id")
	}

	userID, appErr := utils.StrIDToPgUUID(authz[0])
	if appErr != nil || !userID.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid user id")
	}

	authzMeta := request.AuthzMetadata{
		UserID: userID,
	}

	appErr = h.authService.Signout(ctx, req, authzMeta)
	if appErr != nil {
		return nil, status.Errorf(codes.Internal, "%s", appErr.Message)
	}

	return &user_servicev1.SignoutResponse{
		Success: true,
	}, nil
}
