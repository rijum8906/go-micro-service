package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

func (r *authRepository) CreateSession(ctx context.Context, metadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError) {
	refreshToken, err := r.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	session, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		AccountID:    authzMetadata.UserID,
		UserAgent:    metadata.UserAgent,
		DeviceID:     metadata.DeviceID,
		IpAddr:       metadata.IPAddr,
		RefreshToken: refreshToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(r.env.ScopedJwtExpiration),
			Valid: true,
		},
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &session, nil
}

func (r *authRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError) {
	session, err := r.q.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &session, nil
}

func (r *authRepository) RevokeSession(ctx context.Context, id pgtype.UUID, authzMetadata request.AuthzMetadata) *errors.AppError {
	err := r.q.DeleteSession(ctx, id)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (r *authRepository) RevokeAllSessions(ctx context.Context, authzMetadata request.AuthzMetadata) *errors.AppError {
	return nil
}

func (r *authRepository) DeleteExpiredSessions() {}
