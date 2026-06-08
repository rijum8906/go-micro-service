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

func (s *SessionService) RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke session request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
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
	if appErr := s.TokenManager.RevokeAuthToken(ctx, userInfo.TokenID, userInfo.UserID); appErr != nil {
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
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
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
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
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

	var (
		accessToken      string
		refreshTokenHash string
	)

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		currentSession, err := q.GetSession(ctx, currentSessionID)
		if appErr := utils.AssertRowExists(err, "session", userInfo.UserID); appErr != nil {
			return appErr
		}

		if currentSession.UserID.String() != userInfo.UserID {
			return apperror.ErrForbidden.WithMessage("session does not belong to user")
		}

		if err = q.RevokeOtherSessions(ctx, db.RevokeOtherSessionsParams{
			UserID: userID,
			ID:     currentSessionID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("user not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to revoke other sessions").
				WithDetail("db_error", err.Error())
		}

		if appErr := s.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID); appErr != nil {
			return appErr
		}

		hash, appErr := s.HashService.Generate(32)
		if appErr != nil {
			return appErr
		}
		refreshTokenHash = hash

		tokenRes, appErr := s.TokenManager.IssueAuthToken(ctx, userInfo.UserID, userInfo.SessionID)
		if appErr != nil {
			return appErr
		}
		accessToken = tokenRes.TokenString

		if _, err := q.UpdateSessionRefreshTokenHash(ctx, db.UpdateSessionRefreshTokenHashParams{
			ID:               currentSessionID,
			RefreshTokenHash: refreshTokenHash,
		}); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to update session refresh token hash").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &modelsv1.AuthToken{
		AccessToken: &modelsv1.Token{
			Value:     accessToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.SessionTTL),
		},
		RefreshToken: &modelsv1.Token{
			Value:     refreshTokenHash,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.RefreshTokenTTL),
		},
	}, nil
}
