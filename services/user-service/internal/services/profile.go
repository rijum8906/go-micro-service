package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *userService) UpdateProfile(ctx context.Context, data dto.UpdateProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *appError.AppError) {
	// 1. Convert the ID
	pgtypeProfileID, appErr := ToPgUUID(data.ProfileID)
	if appErr != nil {
		return nil, appErr
	}

	if data.Avatar != nil {
		file, err := data.Avatar.Open()
		if err != nil {
			return nil, appError.NewAppError(http.StatusInternalServerError, "failed to upload avatar", []appError.Error{
				{Field: "avatar", Message: err.Error()},
			})
		}
		defer file.Close()
		url, err := s.utilsConfig.Storage.UploadFile(ctx, fmt.Sprintf("%s/avatar_url", authzMetadata.UserID), file, data.Avatar.Header.Get("Content-Type"))
		if err != nil {
			return nil, appError.NewAppError(http.StatusInternalServerError, "failed to upload avatar", []appError.Error{
				{Field: "avatar", Message: err.Error()},
			})
		}
		data.AvatarURL = &url
	}
	firstName := pgtype.Text{
		String: getStringValue(data.FirstName),
		Valid:  data.FirstName != nil && *data.FirstName != "",
	}

	lastName := pgtype.Text{
		String: getStringValue(data.LastName),
		Valid:  data.LastName != nil && *data.LastName != "",
	}

	// Use the URL from your MinIO upload result here
	avatarURL := pgtype.Text{
		String: getStringValue(data.AvatarURL),
		Valid:  data.AvatarURL != nil && *data.AvatarURL != "",
	}

	updateProfileParams := db.UpdateProfileParams{
		ID: pgtypeProfileID,
	}
	if avatarURL.Valid {
		updateProfileParams.AvatarUrl = avatarURL
	}
	if firstName.Valid {
		updateProfileParams.FirstName = firstName
	}
	if lastName.Valid && getStringValue(data.AvatarURL) != "" {
		updateProfileParams.LastName = lastName
	}

	// 2. Execute the update with nullable parameters
	profile, err := s.q.UpdateProfile(ctx, updateProfileParams)
	if err != nil {
		// Log the actual error internally before returning a generic AppError
		return nil, appError.NewAppError(http.StatusInternalServerError, "internal server error", nil)
	}

	return &profile, nil
}
