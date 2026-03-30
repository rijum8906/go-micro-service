package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

const (
	defaultSessionPage  int32 = 1
	defaultSessionLimit int32 = 20
	activeSessionLimit  int32 = 100
)

func (s *sessionService) GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest) (*sessionv1.GetSessionsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get sessions request is required")
	}

	userID, appErr := parseSessionUserID(req.GetUserId())
	if appErr != nil {
		return nil, appErr
	}

	page, limit := normalizePagination(req.GetPage())
	offset := (page - 1) * limit

	sessions, appErr := s.repos.Session.GetActiveSessions(ctx, userID, limit, offset)
	if appErr != nil {
		return nil, appErr
	}

	return &sessionv1.GetSessionsResponse{
		Sessions: utils.MapSessions(*sessions),
	}, nil
}

func (s *sessionService) GetActiveSessions(ctx context.Context, req *sessionv1.GetActiveSessionsRequest) (*sessionv1.GetActiveSessionsResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get active sessions request is required")
	}

	userID, appErr := parseSessionUserID(req.GetUserId())
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

func parseSessionUserID(userID string) (uuid.UUID, *apperror.AppError) {
	if userID == "" {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("user id is required")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	return parsedUserID, nil
}

func normalizePagination(pageReq interface {
	GetPage() int32
	GetLimit() int32
}) (int32, int32) {
	page := pageReq.GetPage()
	if page <= 0 {
		page = defaultSessionPage
	}

	limit := pageReq.GetLimit()
	if limit <= 0 {
		limit = defaultSessionLimit
	}

	return page, limit
}
