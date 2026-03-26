// Package handler exposes the auth gRPC server implementation.
package handler

import (
	"context"

	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/dto"
	authservice "github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/utils"
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
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	result, appErr := h.service.Register(ctx, dto.Register{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}, &dto.RequestMeta{
		DeviceID: req.GetMetadata().GetDeviceId(),
	})
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
