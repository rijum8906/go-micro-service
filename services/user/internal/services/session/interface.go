// Package session
package session

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/metadata"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type SessionService interface {
	GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest) (*sessionv1.GetSessionsResponse, *apperror.AppError)
	GetActiveSessions(ctx context.Context, userInfo *metadata.UserInfo) (*sessionv1.GetActiveSessionsResponse, *apperror.AppError)
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
