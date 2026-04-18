// Package auth
package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, data *authv1.LoginRequest, client *dto.ClientInfo) (*authv1.AuthResponse, *apperror.AppError)
	Register(ctx context.Context, data *authv1.RegisterRequest, client *dto.ClientInfo) (*authv1.AuthResponse, *apperror.AppError)
	Logout(ctx context.Context, client *dto.UserInfo) (bool, *apperror.AppError)
	RefreshToken(ctx context.Context, user *dto.UserInfo) (*authv1.RefreshTokenResponse, *apperror.AppError)
	RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*corev1.SuccessResponse, *apperror.AppError)
	RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*corev1.SuccessResponse, *apperror.AppError)
	VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*corev1.SuccessResponse, *apperror.AppError)
	ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*corev1.SuccessResponse, *apperror.AppError)
}

type authService struct {
	env       *coreenv.CoreEnv
	repos     *utils.Repos
	utils     *utils.ServiceUtils
	publisher *broker.Client
}

func NewAuthService(repo *utils.Repos, utils *utils.ServiceUtils, env *coreenv.CoreEnv, publisher *broker.Client) (AuthService, *apperror.AppError) {
	if repo == nil || repo.User == nil || repo.Profile == nil || repo.Session == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("repos", "auth repositories are not configured")
	}
	if utils == nil || utils.TokenManager == nil || utils.HashService == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("utils", "auth utilities are not configured")
	}
	if env == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("env", "auth environment config is not configured")
	}
	if publisher == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize auth service").WithDetail("publisher", "job publisher is not configured")
	}

	return &authService{
		env:       env,
		repos:     repo,
		utils:     utils,
		publisher: publisher,
	}, nil
}
