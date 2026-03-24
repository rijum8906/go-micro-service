// Package profile
package profile

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*db.Profile, *apperror.AppError)
	GetProfileByUserID(ctx context.Context, id uuid.UUID) (*db.Profile, *apperror.AppError)
	UpdateProfileNames(ctx context.Context, id uuid.UUID, firstName, lastName string) (*db.Profile, *apperror.AppError)
	UpdateProfileAvatar(ctx context.Context, id uuid.UUID, avatar string) (*db.Profile, *apperror.AppError)
	DeleteProfile(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type profileRepository struct {
	q db.Querier
}

func NewProfileRepository(q db.Querier) ProfileRepository {
	return &profileRepository{
		q: q,
	}
}
