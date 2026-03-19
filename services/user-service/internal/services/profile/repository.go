package profile

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	profilev1 "github.com/rijum8906/relay/packages/pb/user_service/profile/v1"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

func (r *profileRepository) CreateProfile(ctx context.Context, accountID pgtype.UUID, data *authv1.SignupRequest) (*db.Profile, *errors.AppError) {
	profile, err := r.q.CreateProfile(ctx, db.CreateProfileParams{
		AccountID: accountID,
		FirstName: data.FirstName.Value,
		LastName:  data.LastName.Value,
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &profile, nil
}

func (r *profileRepository) GetProfilesByAccountID(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError) {
	profiles, err := r.q.GetProfilesByAccountID(ctx, accountID)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return &profiles, nil
}

func (r *profileRepository) GetProfile(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError) {
	profile, err := r.q.GetProfile(ctx, profileID)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return &profile, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError) {
	profile, err := r.q.UpdateProfile(ctx, db.UpdateProfileParams{
		ID: profileID,
		FirstName: pgtype.Text{
			String: data.FirstName.Value,
			Valid:  data.FirstName.ProtoReflect().IsValid(),
		},
		LastName: pgtype.Text{
			String: data.LastName.Value,
			Valid:  data.LastName.ProtoReflect().IsValid(),
		},
		DisplayName: pgtype.Text{
			String: data.DisplayName.Value,
			Valid:  data.DisplayName.ProtoReflect().IsValid(),
		},
		AvatarUrl: pgtype.Text{
			String: data.AvatarUrl.Value,
			Valid:  data.AvatarUrl.ProtoReflect().IsValid(),
		},
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return &profile, nil
}
