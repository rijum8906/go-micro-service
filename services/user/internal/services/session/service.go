package session

import (
	"context"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

const (
	activeSessionLimit int32 = 100
)

func (s *sessionService) GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest, userInfo *metadata.UserInfo) (*sessionv1.GetSessionsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get sessions request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	
	
	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	
	offset := (req.GetPage().GetPage() - 1) * req.GetPage().GetLimit()

	sessions, appErr := s.repos.Session.GetActiveSessions(ctx, userID, req.GetPage().GetLimit(), offset)
	if appErr != nil {
		return nil, appErr
	}

	return &sessionv1.GetSessionsResponse{
		Sessions: utils.MapSessions(*sessions),
	}, nil
	
}

func (s *sessionService) GetActiveSessions(ctx context.Context, userInfo *metadata.UserInfo) (*sessionv1.GetActiveSessionsResponse, *apperror.AppError) {
	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	sessions, appErr := s.repos.Session.GetActiveSessions(ctx, userID, activeSessionLimit, 0)
	if appErr != nil {
		return nil, appErr
	}

	return &sessionv1.GetActiveSessionsResponse{
		Sessions: utils.MapActiveSessions(*sessions, time.Now()),
	}, nil
}
