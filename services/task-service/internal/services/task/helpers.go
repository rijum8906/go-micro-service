package task

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	coremetadata "github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func userInfoFromContext(ctx context.Context) (*coredto.UserInfo, *apperror.AppError) {
	userInfo, ok := coremetadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithMessage("user metadata is required")
	}

	if strings.TrimSpace(userInfo.UserID) == "" {
		return nil, apperror.ErrUnAuthenticated.WithMessage("user metadata is required")
	}

	return &userInfo, nil
}

func withUserInfo[T any](ctx context.Context, fn func(*coredto.UserInfo) (T, *apperror.AppError)) (T, error) {
	var zero T

	userInfo, appErr := userInfoFromContext(ctx)
	if appErr != nil {
		return zero, utils.MapAppError(appErr)
	}

	res, appErr := fn(userInfo)
	if appErr != nil {
		return zero, utils.MapAppError(appErr)
	}

	return res, nil
}

func requiredUUID(value, field, requiredMessage string) (uuid.UUID, *apperror.AppError) {
	if strings.TrimSpace(value) == "" {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage(requiredMessage)
	}

	id, appErr := utils.NewUUID(value)
	if appErr != nil {
		return uuid.UUID{}, appErr.WithDetail("field", field)
	}

	return id, nil
}
