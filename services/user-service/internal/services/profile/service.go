package profile

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/response"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func assertString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func (s *profileService) GetProfile(ctx context.Context, id string) (*response.GetProfileResult, *errors.AppError) {
	pgID, appErr := utils.StrIDToPgUUID(id)
	if appErr != nil {
		return nil, appErr
	}

	profile, err := s.q.GetProfile(ctx, pgID)
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &response.GetProfileResult{
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		DisplayName: &profile.DisplayName.String,
		AvatarURL:   &profile.AvatarUrl.String,
	}, nil
}

func (s *profileService) UpdateProfile(
	ctx context.Context,
	data request.UpdateProfileRequest,
	reqMetadata request.RequestMetadata,
	authzMetadata request.AuthzMetadata,
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
	pgID, appErr := utils.StrIDToPgUUID(data.ProfileID)
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
		DisplayName: pgtype.Text{
			String: assertString(data.DisplayName),
			Valid:  data.DisplayName != nil,
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
	data request.CreateProfileRequest,
	reqMetadata request.RequestMetadata,
	authzMetadata request.AuthzMetadata,
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
	id string,
	authzMetadata request.AuthzMetadata,
) *errors.AppError {
	pgID, appErr := utils.StrIDToPgUUID(id)
	if appErr != nil {
		return appErr
	}
	err := s.q.DeleteProfile(ctx, pgID)
	if err != nil {
		return errors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (s *profileService) MyProfile(ctx context.Context, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Profile, *errors.AppError) {
	profile, err := s.q.GetProfile(ctx, authzMetadata.UserID)
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}
	return &profile, nil
}
