package session

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *SessionService) GetSessions(ctx context.Context, req *corev1.PaginationRequest) (*sessionv1.GetSessionsResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get sessions request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id")
	}

	offset := (req.GetPage() - 1) * req.GetLimit()

	sessions, err := s.DBQ.GetActiveSessionsByUserID(ctx, db.GetActiveSessionsByUserIDParams{
		UserID: userID,
		Limit:  req.GetLimit(),
		Offset: offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get active sessions").WithDetail("error", err.Error())
	}

	mappedSessions := make([]*modelsv1.Session, len(sessions))
	for i := range sessions {
		mappedSessions[i] = utils.MapSession(&sessions[i])
	}

	return &sessionv1.GetSessionsResponse{
		Sessions: mappedSessions,
	}, nil
}

func (s *SessionService) GetActiveSessions(ctx context.Context, req *corev1.PaginationRequest) (*sessionv1.GetActiveSessionsResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get sessions request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id")
	}

	sessions, err := s.DBQ.GetActiveSessionsByUserID(ctx, db.GetActiveSessionsByUserIDParams{
		UserID: userID,
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get active sessions").WithDetail("error", err.Error())
	}

	activeSessions := make([]*modelsv1.Session, len(sessions))
	for i := range sessions {
		activeSessions[i] = utils.MapSession(&sessions[i])
	}

	return &sessionv1.GetActiveSessionsResponse{
		Sessions: activeSessions,
	}, nil
}

func (s *SessionService) GetCurrentSession(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.Session, error) {
	// validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("get current session request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	sessionID, err := uuid.Parse(userInfo.SessionID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid session id")
	}

	session, err := s.DBQ.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("session not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get session by id").
			WithDetail("db_error", err.Error())
	}

	if session.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrForbidden.WithMessage("session does not belong to user")
	}

	return utils.MapSession(&session), nil
}

func (s *SessionService) RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke session request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	sessionID, err := uuid.Parse(userInfo.SessionID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid session id")
	}

	// Delete refresh token from db
	if err = s.DBQ.RevokeSession(ctx, sessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("session not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to revoke session").
			WithDetail("db_error", err.Error())
	}

	// delete access token from cache
	if appErr := s.TokenManager.RevokeAuthToken(ctx, userInfo.UserID, userInfo.SessionID); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SessionService) RevokeAllSessions(ctx context.Context, req *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke all sessions request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id")
	}

	if err = s.DBQ.RevokeActiveSessions(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("user not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to revoke user sessions").
			WithDetail("db_error", err.Error())
	}

	// delete refresh token from db
	if appErr := s.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SessionService) RevokeOtherSessions(ctx context.Context, req *sessionv1.RevokeOtherSessionsRequest) (*modelsv1.AuthToken, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke other sessions request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id")
	}

	currentSessionID, err := uuid.Parse(userInfo.SessionID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid session id")
	}

	currentSession, err := s.DBQ.GetSession(ctx, currentSessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("session not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get session by id").
			WithDetail("db_error", err.Error())
	}

	if currentSession.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrForbidden.WithMessage("session does not belong to user")
	}

	if err = s.DBQ.RevokeOtherSessions(ctx, db.RevokeOtherSessionsParams{
		UserID: userID,
		ID:     currentSessionID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("user not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to revoke other sessions").
			WithDetail("db_error", err.Error())
	}

	if appErr := s.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID); appErr != nil {
		return nil, appErr
	}

	tokenPair, appErr := s.issueTokenPair(ctx, userInfo.UserID, userInfo.SessionID)
	if appErr != nil {
		return nil, appErr
	}

	return &modelsv1.AuthToken{
		AccessToken: &modelsv1.Token{
			Value:     tokenPair.AccessToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.SessionTTL),
		},
		RefreshToken: &modelsv1.Token{
			Value:     tokenPair.RefreshTokenHash,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.RefreshTokenTTL),
		},
	}, nil
}

// TODO: implement
func (s *SessionService) TerminateExpiredSessions(context.Context, *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	s.Logger.Error("Implement TerminateExpiredSessions")
	return nil, nil
}
