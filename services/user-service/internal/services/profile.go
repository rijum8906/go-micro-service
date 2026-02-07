package services

import (
	"context"
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

func (s *userService) UpdateProfile(ctx context.Context, data dto.UpdateProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) *appError.AppError {
	// 1. Convert the ID
	pgtypeProfileID, appErr := ToPgUUID(data.ProfileID)
	if appErr != nil {
		return appErr
	}

	if data.Avatar != nil {
		file, err := data.Avatar.Open()
		if err != nil {
			return appError.NewAppError(http.StatusInternalServerError, "failed to upload avatar", &[]appError.Error{
				{Field: "avatar", Message: err.Error()},
			})
		}
		defer file.Close()
		url, err := s.utilsConfig.Storage.UploadFile(ctx, data.Avatar.Filename, file, data.Avatar.Header.Get("Content-Type"))
		if err != nil {
			return appError.NewAppError(http.StatusInternalServerError, "failed to upload avatar", &[]appError.Error{
				{Field: "avatar", Message: err.Error()},
			})
		}
		data.AvatarURL = &url
	}
	// 2. Execute the update with nullable parameters
	_, err := s.q.UpdateProfile(ctx, db.UpdateProfileParams{
		ID: pgtypeProfileID,
		FirstName: pgtype.Text{
			String: getStringValue(data.FirstName),
			Valid:  data.FirstName != nil && *data.FirstName != "",
		}.String,
		LastName: pgtype.Text{
			String: getStringValue(data.LastName),
			Valid:  data.LastName != nil && *data.LastName != "",
		}.String,
		AvatarUrl: pgtype.Text{
			String: getStringValue(data.AvatarURL),
			Valid:  data.AvatarURL != nil && *data.AvatarURL != "",
		},
	})
	if err != nil {
		// Log the actual error internally before returning a generic AppError
		return appError.NewAppError(http.StatusInternalServerError, "internal server error", nil)
	}

	return nil
}
