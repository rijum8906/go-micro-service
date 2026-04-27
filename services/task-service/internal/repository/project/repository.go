package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type projectRepository struct {
	q db.Querier
}

func NewProjectRepository(q db.Querier) ProjectRepository {
	return &projectRepository{q: q}
}

func (r *projectRepository) CreateProject(ctx context.Context, params db.CreateProjectParams) (*db.Project, *apperror.AppError) {
	project, err := r.q.CreateProject(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) GetProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	project, err := r.q.GetProject(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("project not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to get project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) UpdateProject(ctx context.Context, params db.UpdateProjectParams) (*db.Project, *apperror.AppError) {
	project, err := r.q.UpdateProject(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("project not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to update project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) CompleteProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	project, err := r.q.CompleteProject(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("project not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to complete project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) ArchiveProject(ctx context.Context, id uuid.UUID) (*db.Project, *apperror.AppError) {
	project, err := r.q.ArchiveProject(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("project not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to archive project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) DeleteProject(ctx context.Context, params db.DeleteProjectParams) (*db.Project, *apperror.AppError) {
	project, err := r.q.DeleteProject(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("project not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to delete project").WithDetail("error", err.Error())
	}

	return &project, nil
}

func (r *projectRepository) ListProjects(ctx context.Context, params db.ListProjectsParams) ([]db.Project, *apperror.AppError) {
	projects, err := r.q.ListProjects(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list projects").WithDetail("error", err.Error())
	}

	return projects, nil
}
