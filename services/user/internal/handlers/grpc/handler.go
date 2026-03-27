// Package handler exposes the auth gRPC server implementation.
package handler

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
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
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("login request is required"))
	}

	metadata, ok := metadata.Receive(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("login request metadata is required"))
	}

	fmt.Println("metadata: ", metadata)

	result, appErr := h.service.Login(ctx, dto.Login{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("register request is required"))
	}

	metadata, ok := metadata.Receive(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("login request metadata is required"))
	}

	result, appErr := h.service.Register(ctx, dto.Register{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
