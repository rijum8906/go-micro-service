package profile

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/db"
)

func (r *profileRepository) CreateProfile(ctx context.Context, params db.CreateProfileParams) (*db.Profile, *apperror.AppError) {
	profile, err := r.q.CreateProfile(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to create profile").WithDetail("error", err.Error())
	}
	return &profile, nil
}

func (r *profileRepository) GetProfile(ctx context.Context, id uuid.UUID) (*db.Profile, *apperror.AppError) {
	profile, err := r.q.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("Profile not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to get profile").WithDetail("error", err.Error())
	}
	return &profile, nil
}

func (r *profileRepository) GetProfileByUserID(ctx context.Context, id uuid.UUID) (*db.Profile, *apperror.AppError) {
	profile, err := r.q.GetProfileByUserID(ctx, id)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to get profile").WithDetail("error", err.Error())
	}
	return &profile, nil
}

func (r *profileRepository) UpdateProfileNames(ctx context.Context, id uuid.UUID, firstName, lastName string) (*db.Profile, *apperror.AppError) {
	profile, err := r.q.UpdateProfileName(ctx, db.UpdateProfileNameParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to update profile").WithDetail("error", err.Error())
	}
	return &profile, nil
}

func (r *profileRepository) UpdateProfileAvatar(ctx context.Context, id uuid.UUID, avatar string) (*db.Profile, *apperror.AppError) {
	profile, err := r.q.UpdateProfileAvatarURL(ctx, db.UpdateProfileAvatarURLParams{
		ID:        id,
		AvatarUrl: avatar,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to update profile").WithDetail("error", err.Error())
	}
	return &profile, nil
}

func (r *profileRepository) DeleteProfile(ctx context.Context, id uuid.UUID) *apperror.AppError {
	err := r.q.DeleteProfile(ctx, id)
	if err != nil {
		return apperror.ErrInternal.WithMessage("Failed to delete profile").WithDetail("error", err.Error())
	}
	return nil
}
