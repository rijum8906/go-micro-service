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

	// 2. Execute the update with nullable parameters
	_, err := s.q.UpdateProfile(ctx, db.UpdateProfileParams{
		ID: pgtypeProfileID,
		FirstName: pgtype.Text{
			String: getStringValue(data.FirstName),
			Valid:  data.FirstName != nil,
		}.String,
		LastName: pgtype.Text{
			String: getStringValue(data.LastName),
			Valid:  data.LastName != nil,
		}.String,
		AvatarUrl: pgtype.Text{
			String: getStringValue(data.AvatarURL),
			Valid:  data.AvatarURL != nil,
		},
	})
	if err != nil {
		// Log the actual error internally before returning a generic AppError
		return appError.NewAppError(http.StatusInternalServerError, "internal server error", nil)
	}

	return nil
}
