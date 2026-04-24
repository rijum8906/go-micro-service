package utils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	taskrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task"
)

type Repos struct {
	Task taskrepo.TaskRepository
}

func NewUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}
	return u, nil
}
