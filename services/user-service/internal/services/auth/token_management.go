package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
)

// RefreshToken refreshes the access token using the provided refresh token
//
// Constraints:
// - The refresh token must be valid and not expired
//
// Race Conditions Prevention:
// - We locked the session to prevent concurrent updates
//
// NOTE: this method does not require authentication
func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshAccessTokenRequest) (*authv1.RefreshTokenResponse, error) {
	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.GetClientInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "client info not found in context")
	}

	var (
		refreshTokenHash string
		session          *db.Session
	)

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Lock the session to prevent concurrent updates
		sess, err := q.LockAndGetSessionByRefreshTokenHash(ctx, req.RefreshToken)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("session not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to lock session").
				WithDetail("db_error", err.Error())
		}
		session = &sess

		// Validate device ID and expiry
		if session.ExpiresAt.Time.Before(time.Now()) {
			return apperror.ErrPermissionDenied.WithMessage("session is already expired")
		}
		if session.DeviceID != clientInfo.DeviceID {
			return apperror.ErrPermissionDenied.WithMessage("device mismatch")
		}

		// Generate a refresh token hash
		hash, appErr := s.HashService.Generate(32)
		if appErr != nil {
			return appErr
		}
		refreshTokenHash = hash

		// Persist tokens in db
		if _, err := q.UpdateSessionRefreshTokenHash(ctx, db.UpdateSessionRefreshTokenHashParams{
			ID:               session.ID,
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

	// Issue new access token and refresh token
	accessToken, appErr := s.issueAccessToken(ctx, session.UserID.String(), session.ID.String())
	if appErr != nil {
		return nil, appErr
	}

	// Return the refreshed token response
	return &authv1.RefreshTokenResponse{
		AccessToken: &modelsv1.Token{
			Value:     accessToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.SessionTTL),
		},
		RefreshToken: refreshTokenHash,
	}, nil
}

// GenerateScopedToken generates a scoped token for the given request
//
// Constraints:
// - User must be authenticated
// - Auth method must be password (implement other auth methods)
// - Scope must be valid
// - Auth value must be valid for the given auth method
func (s *AuthService) GenerateScopedToken(ctx context.Context, req *authv1.GenerateScopedTokenRequest) (*authv1.GenerateScopedTokenResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, constants.ErrInvalidUserIDInUserInfo
	}

	// Validate auth method and token scope
	if req.AuthMethod != string(coreconstants.AuthMethodPassword) {
		return nil, apperror.ErrValidation.WithMessage("invalid auth method")
	}
	if !token.IsValidTokenScope(req.GetScope()) {
		return nil, apperror.ErrValidation.WithMessage("invalid token scope")
	}
	if !constants.IsInternalTokenScope(req.GetScope()) {
		return nil, apperror.ErrValidation.WithMessage("invalid token scope for generate scoped token")
	}

	// Retrieve user from database
	user, err := s.DBQ.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Do a dummy hashing to avoid timing attacks
			s.HashService.Hash(req.GetAuthValue())
			return nil, apperror.ErrNotFound.WithMessage("user not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get user by id").
			WithDetail("db_error", err.Error())
	}

	// Verify password
	if !s.HashService.Verify(user.PasswordHash.String, req.AuthValue) {
		return nil, apperror.ErrValidation.WithMessage("invalid password")
	}

	// Generate scoped token
	tokenRes, appErr := s.TokenManager.GenerateToken(
		userInfo.UserID,
		uuid.NewString(),
		req.GetScope(),
		s.Config.ScopedTokenTTL,
	)
	if appErr != nil {
		return nil, appErr
	}

	// Return token response
	return &authv1.GenerateScopedTokenResponse{
		ScopedToken: tokenRes.TokenString,
	}, nil
}
