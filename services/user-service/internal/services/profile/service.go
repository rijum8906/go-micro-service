package profile

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	profilev1 "github.com/rijum8906/relay/packages/pb/user_service/profile/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func (s *profileService) CreateProfile(ctx context.Context, data *authv1.SignupRequest, authzMetadata *request.AuthzMetadata) (*db.Profile, *errors.AppError) {
	return s.repo.CreateProfile(ctx, authzMetadata.UserID, data)
}

func (s *profileService) GetProfile(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError) {
	return s.repo.GetProfile(ctx, profileID)
}

func (s *profileService) GetProfilesByAccountID(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError) {
	return s.repo.GetProfilesByAccountID(ctx, accountID)
}

func (s *profileService) UpdateProfile(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError) {
	return s.repo.UpdateProfile(ctx, profileID, data)
}

func (s *profileService) UpdateDisplayName(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateDisplayNameRequest) (*db.Profile, *errors.AppError) {
	profile, err := s.repo.UpdateProfile(ctx, profileID, &profilev1.UpdateProfileRequest{
		DisplayName: utils.NewName(data.DisplayName.Value),
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return profile, nil
}

func (s *profileService) UpdateAvatarURL(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateAvatarUrlRequest) (*db.Profile, *errors.AppError) {
	profile, err := s.repo.UpdateProfile(ctx, profileID, &profilev1.UpdateProfileRequest{
		AvatarUrl: utils.NewURL(data.AvatarUrl.Value),
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return profile, nil
}

func (s *profileService) UpdateName(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateNameRequest) (*db.Profile, *errors.AppError) {
	profile, err := s.repo.UpdateProfile(ctx, profileID, &profilev1.UpdateProfileRequest{
		FirstName: utils.NewName(data.FirstName.Value),
		LastName:  utils.NewName(data.LastName.Value),
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return profile, nil
}
