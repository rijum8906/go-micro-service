package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, params db.CreateProjectParams) (*db.Project, *apperror.AppError)
	GetProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError)
	UpdateProject(ctx context.Context, params db.UpdateProjectParams) (*db.Project, *apperror.AppError)
	CompleteProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError)
	ArchiveProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError)
	DeleteProject(ctx context.Context, params db.DeleteProjectParams) (*db.Project, *apperror.AppError)
	ListProjects(ctx context.Context, params db.ListProjectsParams) ([]db.Project, *apperror.AppError)
}
