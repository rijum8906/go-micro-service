// Package session
package session

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/env"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type SessionService interface {
	GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest, userInfo *dto.UserInfo) (*sessionv1.GetSessionsResponse, *apperror.AppError)
	GetActiveSessions(ctx context.Context, userInfo *dto.UserInfo) (*sessionv1.GetActiveSessionsResponse, *apperror.AppError)
	GetCurrentSession(ctx context.Context, userInfo *dto.UserInfo) (*modelsv1.Session, *apperror.AppError)
	RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	RevokeAllSessions(ctx context.Context, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	RevokeOtherSessions(ctx context.Context, req *sessionv1.RevokeOtherSessionsRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	TerminateExpiredSessions(ctx context.Context) (*corev1.SuccessResponse, *apperror.AppError)
  RevokeOtherSessions(ctx context.Context, req *sessionv1.RevokeOtherSessionsRequest, userInfo *metadata.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
}

type sessionService struct {
	env   *env.Config
	repos *utils.Repos
	utils *utils.Utils
}

func NewSessionService(repo *utils.Repos, utils *utils.Utils, env *env.Config) (SessionService, *apperror.AppError) {
	if repo == nil || repo.User == nil || repo.Profile == nil || repo.Session == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize session service").WithDetail("repos", "session repositories are not configured")
	}
	if utils == nil || utils.TokenManager == nil || utils.HashService == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize session service").WithDetail("utils", "session utilities are not configured")
	}
	if env == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize session service").WithDetail("env", "session environment config is not configured")
	}

	return &sessionService{
		env:   env,
		repos: repo,
		utils: utils,
	}, nil
}
