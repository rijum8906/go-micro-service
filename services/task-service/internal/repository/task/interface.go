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
	UpdateTask(ctx context.Context, params db.UpdateTaskParams) (*db.Task, *apperror.AppError) 
	DeleteTask(ctx context.Context, params db.DeleteTaskParams) (*db.Task, *apperror.AppError)
	ArchiveTask(ctx context.Context, params db.ArchiveTaskParams) (*db.Task, *apperror.AppError)
	UpdateTaskStatus(ctx context.Context, params db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError)
    UpdateTaskProgress(ctx context.Context, params db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError)
    ListTasksByOrganization(ctx context.Context, params db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError)
    ListTasksByParent(ctx context.Context, parentTaskID pgtype.UUID) ([]db.Task, *apperror.AppError)
    ListTasksByCreator(ctx context.Context, params db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError)
    
}
