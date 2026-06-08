package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"go.uber.org/zap"
)

// LogoutAllDevices revokes all active tokens for the user, effectively logging them out from all devices.
//
// Constraints:
//   - User must be authenticated
func (s *AuthService) LogoutAllDevices(ctx context.Context, req *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	// Extract user ID from context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, constants.ErrUserNotFoundInCtx
	}
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, constants.ErrInvalidUserIDInUserInfo
	}

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Mark all devices as revoked in the database
		if err := q.RevokeActiveSessions(ctx, userID); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to revoke active sessions in database").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Revoke all devices for the user
	if appErr := s.TokenManager.RevokeAllUserTokens(ctx, userInfo.UserID); appErr != nil {
		s.Logger.Error("failed to revoke all user tokens", apperror.ParseAppErrorIntoZapFields(appErr)...)
	}

	// Why revoke from database before Redis?
	// The database is our ultimate source of truth. By marking sessions revoked here first,
	// any concurrent attempts to refresh access tokens will fail.
	//
	// If the subsequent Redis eviction fails, the session is still dead on the DB level,
	// though stale cache records may persist until their natural TTL expiration.

	// Returns success response
	return &corev1.SuccessResponse{Success: true}, nil
}

// Logout terminates the user's current session and revokes authentication tokens.
//
// Constraints:
// - The user must be authenticated via a valid session and token.
// - The device id must match the device id of the session.
//
// Idempotent:
//   - multiple calls with the same token will not affect the session state.
func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*corev1.SuccessResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, constants.ErrUserNotFoundInCtx
	}
	sessionID, err := uuid.Parse(userInfo.SessionID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to parse session ID")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.GetClientInfoFromIncomingContext(ctx)
	if !ok {
		return nil, constants.ErrClientNotFoundInCtx
	}

	appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Fetch session from database
		session, err := q.GetSession(ctx, sessionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("session not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get session by id").
				WithDetail("db_error", err.Error())
		}
		// Verify device ID matches the client info
		if session.DeviceID != clientInfo.DeviceID {
			return apperror.ErrPermissionDenied.WithMessage("device mismatch")
		}

		// Idempotent revoke session from database
		if session.IsRevoked {
			return nil
		}

		if err := q.RevokeSession(ctx, session.ID); err != nil {
			return apperror.ErrInternal.WithDetail("internal_message", "failed to revoke session").WithDetail("db_error", err.Error())
		}

		return nil
	})
	if appErr != nil {
		return nil, appErr
	}

	// Remove access token from redis
	if appErr := s.TokenManager.RevokeAuthToken(ctx, userInfo.TokenID, userInfo.UserID); appErr != nil {
		s.Logger.Error("failed to revoke access token", apperror.ParseAppErrorIntoZapFields(appErr)...)
		return nil, appErr
	}

	// Log successful logout for audit trail
	s.Logger.Info("user logged out successfully",
		zap.String("user_id", userInfo.UserID),
		zap.String("session_id", userInfo.SessionID))

	return &corev1.SuccessResponse{Success: true}, nil
}
