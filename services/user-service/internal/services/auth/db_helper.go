package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
)

// getUserProfile retrieves the user's profile.
func (s *AuthService) getUserProfile(ctx context.Context, q *db.Queries, userID uuid.UUID) (*db.Profile, *apperror.AppError) {
	profile, err := q.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInternal.
				WithMessage("user profile not found. Please contact support").
				WithDetail("internal_message", "failed to retrieve user profile").
				WithDetail("user_id", userID.String())
		}

		return nil, apperror.ErrInternal.
			WithMessage("failed to retrieve user profile").
			WithDetail("error", err.Error())
	}

	return &profile, nil
}

// createProfile creates a user profile record.
func (s *AuthService) createProfile(ctx context.Context, q *db.Queries, userID uuid.UUID, req *authv1.RegisterRequest) (*db.Profile, *apperror.AppError) {
	profile, err := q.CreateProfile(ctx, db.CreateProfileParams{
		UserID:    userID,
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to create user profile").
			WithDetail("db_error", err.Error())
	}

	return &profile, nil
}

// createSession creates a new session record for the authenticated user.
func (s *AuthService) createSession(ctx context.Context, q *db.Queries, userID uuid.UUID, refreshTokenHash string, clientInfo *dto.ClientInfo) (*db.Session, *apperror.AppError) {
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(s.Config.SessionTTL),
		Valid: true,
	}

	session, err := q.CreateSession(ctx, db.CreateSessionParams{
		UserID:           userID,
		UserAgent:        clientInfo.UserAgent,
		IpAddr:           clientInfo.IPAddress,
		DeviceID:         clientInfo.DeviceID,
		ExpiresAt:        expiresAt,
		RefreshTokenHash: refreshTokenHash,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to create session").
			WithDetail("db_error", err.Error())
	}

	return &session, nil
}
