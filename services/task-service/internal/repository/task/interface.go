package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, params db.CreateTaskParams) (*db.Task, *apperror.AppError)
	GetTask(ctx context.Context, id uuid.UUID) (*db.Task, *apperror.AppError)
	ListTasksByProject(ctx context.Context, projectID pgtype.UUID) ([]db.Task, *apperror.AppError)
}
