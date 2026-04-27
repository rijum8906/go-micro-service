package task

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type taskRepository struct {
	q db.Querier
}

func NewTaskRepository(q db.Querier) TaskRepository {
	return &taskRepository{q: q}
}

func (r *taskRepository) CreateTask(ctx context.Context, params db.CreateTaskParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.CreateTask(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create task").WithDetail("error", err.Error())
	}

	return &task, nil
}

func (r *taskRepository) GetTask(ctx context.Context, id uuid.UUID) (*db.Task, *apperror.AppError) {
	task, err := r.q.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("task not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to get task").WithDetail("error", err.Error())
	}

	return &task, nil
}

func (r *taskRepository) ListTasksByProject(ctx context.Context, projectID pgtype.UUID) ([]db.Task, *apperror.AppError) {
	tasks, err := r.q.ListTasksByProject(ctx, projectID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list tasks by project").WithDetail("error", err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) UpdateTask(ctx context.Context, params db.UpdateTaskParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.UpdateTask(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to update task").WithDetail("error", err.Error())
	}
	
	return &task, nil
}

func (r *taskRepository) DeleteTask(ctx context.Context, params db.DeleteTaskParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.DeleteTask(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to delete task").WithDetail("error", err.Error())
	}
	
	return &task, nil
}

func (r *taskRepository) ArchiveTask(ctx context.Context, params db.ArchiveTaskParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.ArchiveTask(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("task not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to get task").WithDetail("error", err.Error())
	}

	return &task, nil
}

func (r *taskRepository) UpdateTaskStatus(ctx context.Context, params db.UpdateTaskStatusParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.UpdateTaskStatus(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to update task status").WithDetail("error", err.Error())
	}
	
	return &task, nil
}

func (r *taskRepository) UpdateTaskProgress(ctx context.Context, params db.UpdateTaskProgressParams) (*db.Task, *apperror.AppError) {
	task, err := r.q.UpdateTaskProgress(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to update task progress").WithDetail("error", err.Error())
	}
	
	return &task, nil
}

func (r *taskRepository) ListTasksByOrganization(ctx context.Context, params db.ListTasksByOrganizationParams) ([]db.Task, *apperror.AppError) {
	tasks, err := r.q.ListTasksByOrganization(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list tasks by organization").WithDetail("error", err.Error())
	}
	
	return tasks, nil
}

func (r *taskRepository) ListTasksByParent(ctx context.Context, parentTaskID pgtype.UUID) ([]db.Task, *apperror.AppError) {
	tasks, err := r.q.ListTasksByParent(ctx, parentTaskID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list tasks by parent").WithDetail("error", err.Error())
	}
	
	return tasks, nil
}

func (r *taskRepository) ListTasksByCreator(ctx context.Context, params db.ListTasksByCreatorParams) ([]db.Task, *apperror.AppError) {
	tasks, err := r.q.ListTasksByCreator(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list tasks by creator ").WithDetail("error", err.Error())
	}
	
	return tasks, nil
}
