package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func assertString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func (s *profileService) GetProfile(ctx context.Context, id string) (*dto.GetProfileResult, *errors.AppError) {
	pgID, appErr := ToPgUUID(id)
	if appErr != nil {
		return nil, appErr
	}

	profile, err := s.q.GetProfile(ctx, pgID)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &dto.GetProfileResult{
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		DisplayName: &profile.DisplayName.String,
		AvatarURL:   &profile.AvatarUrl.String,
	}, nil
}

func (s *profileService) UpdateProfile(
	ctx context.Context,
	data dto.UpdateProfileRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) (*db.Profile, *errors.AppError) {
	// Check if the user uploaded avatar
	if data.Avatar != nil {
		// TODO: upload avatar
		file, err := data.Avatar.Open()
		if err != nil {
			return nil, errors.ErrInternal.WithInternal(err)
		}
		defer file.Close()
		publicURL, err := s.utilsConfig.Storage.UploadFile(ctx, fmt.Sprintf("%s/avatar_url", authzMetadata.UserID), file, data.Avatar.Header.Get("Content-Type"))
		if err != nil {
			return nil, errors.ErrInternal.WithInternal(err)
		}
		data.AvatarURL = &publicURL
	}
	pgID, appErr := ToPgUUID(data.ProfileID)
	if appErr != nil {
		return nil, appErr
	}

	updateProfileParams := db.UpdateProfileParams{
		ID: pgID,
		FirstName: pgtype.Text{
			String: assertString(data.FirstName),
			Valid:  data.FirstName != nil,
		},
		LastName: pgtype.Text{
			String: assertString(data.LastName),
			Valid:  data.LastName != nil,
		},
		AvatarUrl: pgtype.Text{
			String: assertString(data.AvatarURL),
			Valid:  data.AvatarURL != nil,
		},
	}

	profile, err := s.q.UpdateProfile(ctx, updateProfileParams)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &profile, nil
}

func (s *profileService) CreateProfile(
	ctx context.Context,
	data dto.CreateProfileRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) (*db.Profile, *errors.AppError) {
	if data.Avatar != nil {
		file, err := data.Avatar.Open()
		if err != nil {
			return nil, errors.ErrInternal.WithInternal(err)
		}

		defer file.Close()
		publicURL, err := s.utilsConfig.Storage.UploadFile(ctx, fmt.Sprintf("%s/avatar_url", authzMetadata.UserID), file, data.Avatar.Header.Get("Content-Type"))
		if err != nil {
			return nil, errors.ErrInternal.WithInternal(err)
		}
		data.AvatarURL = &publicURL
	}
	profile, err := s.q.CreateProfile(ctx, db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		DisplayName: pgtype.Text{
			String: assertString(data.DisplayName),
			Valid:  data.DisplayName != nil,
		},
		AvatarUrl: pgtype.Text{
			String: assertString(data.AvatarURL),
			Valid:  data.AvatarURL != nil,
		},
		AccountID: authzMetadata.UserID,
	})
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}
	return &profile, nil
}

func (s *profileService) DeleteProfile(
	ctx context.Context,
	data dto.DeleteProfileRequest,
	reqMetadata dto.RequestMetadata,
	authzMetadata dto.AuthzMetadata,
) *errors.AppError {
	err := s.q.DeleteProfile(ctx, authzMetadata.UserID)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (s *profileService) MyProfile(ctx context.Context, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *errors.AppError) {
	profile, err := s.q.GetProfile(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}
	return &profile, nil
}
