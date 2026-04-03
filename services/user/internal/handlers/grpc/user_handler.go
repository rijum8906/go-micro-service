package handler

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	userservice "github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	service userservice.UserService
}

func NewUserHandler(service userservice.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GenerateScopedToken(ctx context.Context, req *userv1.GenerateScopedTokenRequest) (*userv1.GenerateScopedTokenResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("generate scoped token request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("generate scoped token user metadata is required"))
	}

	result, appErr := h.service.GenerateScopedToken(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("change password request is required"))
	}

	result, appErr := h.service.ChangePassword(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *UserHandler) UpdateProfileName(ctx context.Context, req *userv1.UpdateProfileNameRequest) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update profile name request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("update profile name user metadata is required"))
	}

	result, appErr := h.service.UpdateProfileName(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *UserHandler) UpdateProfileAvatarUrl(ctx context.Context, req *userv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update profile avatar request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("update profile avatar user metadata is required"))
	}

	result, appErr := h.service.UpdateProfileAvatarUrl(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get profile request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("get profile user metadata is required"))
	}

	result, appErr := h.service.GetProfile(ctx, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.User, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get user request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("get user metadata is required"))
	}

	result, appErr := h.service.GetUser(ctx, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
