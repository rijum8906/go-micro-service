package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
)

func (r *sessionRepository) CreateSession(ctx context.Context, params db.CreateSessionParams) (*db.Session, *apperror.AppError) {
	session, err := r.q.CreateSession(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to create session").WithDetail("error", err.Error())
	}
	return &session, nil
}

func (r *sessionRepository) GetSession(ctx context.Context, id uuid.UUID) (*db.Session, *apperror.AppError) {
	session, err := r.q.GetSession(ctx, id)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get session").WithDetail("error", err.Error())
	}
	return &session, nil
}

func (r *sessionRepository) GetSessionByRefreshToken(ctx context.Context, token string) (*db.Session, *apperror.AppError) {
	session, err := r.q.GetSessionByRefreshTokenHash(ctx, token)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get session").WithDetail("error", err.Error())
	}
	return &session, nil
}

func (r *sessionRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID, limit, offfset int32) (*[]db.Session, *apperror.AppError) {
	sessions, err := r.q.GetActiveSessionsByUserID(ctx, db.GetActiveSessionsByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offfset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get active sessions").WithDetail("error", err.Error())
	}
	return &sessions, nil
}

func (r *sessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) *apperror.AppError {
	err := r.q.RevokeSession(ctx, id)
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to revoke session").WithDetail("error", err.Error())
	}
	return nil
}

func (r *sessionRepository) RevokeAllSessions(ctx context.Context, userID uuid.UUID) *apperror.AppError {
	err := r.q.RevokeActiveSessions(ctx, userID)
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to revoke all sessions").WithDetail("error", err.Error())
	}
	return nil
}

func (r *sessionRepository) RevokeOtherSessions(ctx context.Context, userID, currentSessionID uuid.UUID) *apperror.AppError {
	err := r.q.RevokeOtherSessions(ctx, db.RevokeOtherSessionsParams{
		UserID: userID,
		ID:     currentSessionID,
	})
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to revoke other sessions").WithDetail("error", err.Error())
	}
	return nil
}

func (r *sessionRepository) TerminateExpiredSessions(ctx context.Context) *apperror.AppError {
	err := r.q.DeleteExpiredSessions(ctx)
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to terminate expired sessions").WithDetail("error", err.Error())
	}
	return nil
}
