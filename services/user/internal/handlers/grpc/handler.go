// Package grpc exposes the auth gRPC server implementation.
package grpc

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
	"github.com/rijum8906/relay/services/user/internal/dto"
	authservice "github.com/rijum8906/relay/services/user/internal/services/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	service authservice.AuthService
}

func NewAuthHandler(service authservice.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	result, appErr := h.service.Login(ctx, dto.Login{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}, &dto.RequestMeta{
		DeviceID: req.GetMetadata().GetDeviceId(),
	})
	if appErr != nil {
		return nil, mapAppError(appErr)
	}

	return &authv1.AuthResponse{
		Account: &modelsv1.Account{
			Id:    result.User.ID.String(),
			Email: result.User.Email,
		},
		Profiles: []*modelsv1.Profile{
			{
				FirstName: result.Profile.FirstName,
				LastName:  result.Profile.LastName,
			},
		},
		Tokens: &modelsv1.AuthToken{
			AccessToken:  result.Tokens.AccessToken.Value,
			RefreshToken: result.Tokens.RefreshToken.Value,
		},
	}, nil
}

func mapAppError(appErr *apperror.AppError) error {
	if appErr == nil {
		return nil
	}

	switch appErr.Type {
	case apperror.TypeValidation:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case apperror.TypeUnAuthenticated:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperror.TypeForbidden:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case apperror.TypeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case apperror.TypeThirdParty:
		return status.Error(codes.Unavailable, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
