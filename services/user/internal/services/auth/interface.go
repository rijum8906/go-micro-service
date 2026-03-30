// Package auth
package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/metadata"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, req *authv1.LoginRequest, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError)
	Register(ctx context.Context, req *authv1.RegisterRequest, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError)
	Logout(ctx context.Context, client *metadata.UserInfo) (bool, *apperror.AppError)
	GenerateScopedToken(ctx context.Context, req *authv1.GenerateScopedTokenRequest, user *metadata.UserInfo) (*authv1.GenerateScopedTokenResponse, *apperror.AppError)
	ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, *apperror.AppError)
	UpdateProfileName(ctx context.Context, req *authv1.UpdateProfileNameRequest) (*modelsv1.Profile, *apperror.AppError)
	UpdateProfileAvatarUrl(ctx context.Context, req *authv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, *apperror.AppError)
}

type authService struct {
	env   *env.Config
	repos *utils.Repos
	utils *utils.Utils
}

func NewAuthService(repo *utils.Repos, utils *utils.Utils, env *env.Config) (AuthService, *apperror.AppError) {
	if repo == nil || repo.User == nil || repo.Profile == nil || repo.Session == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("repos", "auth repositories are not configured")
	}
	if utils == nil || utils.TokenManager == nil || utils.HashService == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("utils", "auth utilities are not configured")
	}
	if env == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("env", "auth environment config is not configured")
	}

	return &authService{
		env:   env,
		repos: repo,
		utils: utils,
	}, nil
}
