// Package session
package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
)

type SessionRepository interface {
	GetSession(ctx context.Context, id uuid.UUID) (*db.Session, *apperror.AppError)
	GetActiveSessions(ctx context.Context, userID uuid.UUID, limit, offfset int32) (*[]db.Session, *apperror.AppError)
	RevokeSession(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type sessionRepository struct {
	q db.Querier
}

func NewSessionRepository(q db.Querier) SessionRepository {
	return &sessionRepository{
		q: q,
	}
}
