// Package handler exposes the auth gRPC server implementation.
package handler

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
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

	metadata, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("login request metadata is required"))
	}

	result, appErr := h.service.Login(ctx, req, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("register request is required"))
	}

	metadata, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("login request metadata is required"))
	}

	result, appErr := h.service.Register(ctx, req, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("logout request is required"))
	}

	metadata, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		// NOTE: if not ok then it's the gateway's bug
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("logout user metadata is required"))
	}

	result, appErr := h.service.Logout(ctx, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return &corev1.SuccessResponse{
		Success: result,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("refresh token request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("refresh token metadata is required"))
	}

	result, appErr := h.service.RefreshToken(ctx, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("request email verification request is required"))
	}

	result, appErr := h.service.RequestEmailVerification(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("request password reset request is required"))
	}

	result, appErr := h.service.RequestPasswordReset(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("verify email request is required"))
	}

	result, appErr := h.service.VerifyEmail(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("reset password request is required"))
	}

	result, appErr := h.service.ResetPassword(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
