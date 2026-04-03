package session

import (
	"context"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

const (
	activeSessionLimit int32 = 100
)

func (s *sessionService) GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest, userInfo *dto.UserInfo) (*sessionv1.GetSessionsResponse, *apperror.AppError) {
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

func (s *sessionService) GetActiveSessions(ctx context.Context, userInfo *dto.UserInfo) (*sessionv1.GetActiveSessionsResponse, *apperror.AppError) {
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

func (s *sessionService) GetCurrentSession(ctx context.Context, userInfo *dto.UserInfo) (*modelsv1.Session, *apperror.AppError) {
	if userInfo == nil || userInfo.UserID == "" || userInfo.SessionID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	sessionID, appErr := utils.NewUUID(userInfo.SessionID)
	if appErr != nil {
		return nil, appErr
	}

	session, appErr := s.repos.Session.GetSession(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	if session.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrForbidden.WithMessage("session does not belong to user")
	}

	return utils.MapSession(*session), nil
}

func (s *sessionService) RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke session request is required")
	}
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if req.GetTokenToRevoke() == "" {
		return nil, apperror.ErrValidation.WithMessage("session id is required")
	}

	session, appErr := s.repos.Session.GetSessionByRefreshToken(ctx, req.GetTokenToRevoke())
	if appErr != nil {
		return nil, appErr
	}

	if session.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrForbidden.WithMessage("session does not belong to user")
	}

	// Delete refresh token from db
	appErr = s.repos.Session.RevokeSession(ctx, session.ID)
	if appErr != nil {
		return nil, appErr
	}

	// delete access token from cache
	s.utils.TokenManager.RevokeAuthToken(ctx, userInfo.UserID, session.ID.String())

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *sessionService) RevokeAllSessions(ctx context.Context, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if userInfo == nil || userInfo.UserID == "" || userInfo.SessionID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.repos.Session.RevokeAllSessions(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	// delete refresh token from db
	appErr = s.utils.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	// delete access token from cache
	s.utils.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID)

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *sessionService) RevokeOtherSessions(ctx context.Context, req *sessionv1.RevokeOtherSessionsRequest, userInfo *metadata.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke other sessions request is required")
	}
	if userInfo == nil || userInfo.UserID == "" || userInfo.SessionID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}
	if req.GetCurrentSessionId() == "" {
		return nil, apperror.ErrValidation.WithMessage("current session id is required")
	}
	if req.GetCurrentSessionId() != userInfo.SessionID {
		return nil, apperror.ErrForbidden.WithMessage("current session does not match authenticated session")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	currentSessionID, appErr := utils.NewUUID(req.GetCurrentSessionId())
	if appErr != nil {
		return nil, appErr
	}

	currentSession, appErr := s.repos.Session.GetSession(ctx, currentSessionID)
	if appErr != nil {
		return nil, appErr
	}

	if currentSession.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrForbidden.WithMessage("session does not belong to user")
	}

	appErr = s.repos.Session.RevokeOtherSessions(ctx, userID, currentSessionID)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.utils.TokenManager.RevokeOtherUserTokens(ctx, userInfo.UserID, currentSessionID.String())
	if appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *sessionService) TerminateExpiredSessions(ctx context.Context) (*corev1.SuccessResponse, *apperror.AppError) {
	appErr := s.repos.Session.TerminateExpiredSessions(ctx)
	if appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
