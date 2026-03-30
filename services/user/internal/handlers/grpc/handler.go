// Package handler exposes the auth gRPC server implementation.
package handler

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
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

	metadata, ok := metadata.ReceiveClientInfo(ctx)
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

	metadata, ok := metadata.ReceiveClientInfo(ctx)
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

func (h *AuthHandler) GenerateScopedToken(ctx context.Context, req *authv1.GenerateScopedTokenInput) (*authv1.ScopedTokenResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("generate scoped token request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("generate scoped token user metadata is required"))
	}

	result, appErr := h.service.GenerateScopedToken(ctx, dto.GenerateScopedToken{
		Scope: req.GetScope(),
	}, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *authv1.ChangePasswordInput) (*authv1.MutationResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("change password request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("change password user metadata is required"))
	}

	result, appErr := h.service.ChangePassword(ctx, dto.ChangePassword{
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	}, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) UpdateProfileName(ctx context.Context, req *authv1.UpdateProfileNameInput) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update profile name request is required"))
	}

	result, appErr := h.service.UpdateProfileName(ctx, dto.UpdateProfileName{
		ProfileID: req.GetProfileId(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *AuthHandler) UpdateProfileAvatarUrl(ctx context.Context, req *authv1.UpdateProfileAvatarUrlInput) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update profile avatar request is required"))
	}

	result, appErr := h.service.UpdateProfileAvatarUrl(ctx, dto.UpdateProfileAvatarUrl{
		ProfileID: req.GetProfileId(),
		AvatarURL: req.GetAvatarUrl(),
	})
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
