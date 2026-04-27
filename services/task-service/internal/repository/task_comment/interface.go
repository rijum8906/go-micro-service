package taskcomment

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type TaskCommentRepository interface {
	CreateTaskComment(ctx context.Context, params db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	GetTaskComment(ctx context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError)
	UpdateTaskComment(ctx context.Context, params db.UpdateTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	DeleteTaskComment(ctx context.Context, params db.DeleteTaskCommentParams) (*db.TaskComment, *apperror.AppError)
	ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]db.TaskComment, *apperror.AppError)
}
