package utils

import (
	"strings"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

func NewUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}
	return u, nil
}

func ValidateUserInfo(userInfo *dto.UserInfo) (uuid.UUID, *apperror.AppError) {
	if userInfo == nil || strings.TrimSpace(userInfo.UserID) == "" {
		return uuid.UUID{}, apperror.ErrUnAuthenticated.WithMessage("user metadata is required")
	}
	userID, appErr := NewUUID(userInfo.UserID)
	if appErr != nil {
		return uuid.UUID{}, appErr.WithDetail("field", "user_id")
	}

	return userID, nil
}
