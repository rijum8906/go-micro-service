// Package user
package user

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type UserService interface {
	GenerateScopedToken(ctx context.Context, req *userv1.GenerateScopedTokenRequest, user *metadata.UserInfo) (*userv1.GenerateScopedTokenResponse, *apperror.AppError)
	ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*corev1.SuccessResponse, *apperror.AppError)
	UpdateProfileName(ctx context.Context, req *userv1.UpdateProfileNameRequest) (*modelsv1.Profile, *apperror.AppError)
	UpdateProfileAvatarUrl(ctx context.Context, req *userv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, *apperror.AppError)
	GetProfile(ctx context.Context, user *metadata.UserInfo) (*modelsv1.Profile, *apperror.AppError)
	GetUser(ctx context.Context, user *metadata.UserInfo) (*modelsv1.User, *apperror.AppError)
}

type userService struct {
	env   *env.Config
	repos *utils.Repos
	utils *utils.Utils
}

func NewUserService(repo *utils.Repos, utils *utils.Utils, env *env.Config) (UserService, *apperror.AppError) {
	if repo == nil || repo.User == nil || repo.Profile == nil || repo.Session == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize user service").WithDetail("repos", "user repositories are not configured")
	}
	if utils == nil || utils.TokenManager == nil || utils.HashService == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize user service").WithDetail("utils", "user utilities are not configured")
	}
	if env == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize user service").WithDetail("env", "user environment config is not configured")
	}

	return &userService{
		env:   env,
		repos: repo,
		utils: utils,
	}, nil
}
