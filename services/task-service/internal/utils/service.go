package utils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func NewUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}
	return u, nil
}
