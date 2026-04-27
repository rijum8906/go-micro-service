package taskcomment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)

type taskCommentRepository struct {
	q db.Querier
}

func NewTaskCommentRepository(q db.Querier) TaskCommentRepository {
	return &taskCommentRepository{q: q}
}

func (r *taskCommentRepository) CreateTaskComment(ctx context.Context, params db.CreateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	comment, err := r.q.CreateTaskComment(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create task comment").WithDetail("error", err.Error())
	}

	return &comment, nil
}

func (r *taskCommentRepository) GetTaskComment(ctx context.Context, id uuid.UUID) (*db.TaskComment, *apperror.AppError) {
	comment, err := r.q.GetTaskComment(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("task comment not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to get task comment").WithDetail("error", err.Error())
	}

	return &comment, nil
}

func (r *taskCommentRepository) UpdateTaskComment(ctx context.Context, params db.UpdateTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	comment, err := r.q.UpdateTaskComment(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("task comment not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to update task comment").WithDetail("error", err.Error())
	}

	return &comment, nil
}

func (r *taskCommentRepository) DeleteTaskComment(ctx context.Context, params db.DeleteTaskCommentParams) (*db.TaskComment, *apperror.AppError) {
	comment, err := r.q.DeleteTaskComment(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("task comment not found")
		}

		return nil, apperror.ErrInternal.WithMessage("failed to delete task comment").WithDetail("error", err.Error())
	}

	return &comment, nil
}

func (r *taskCommentRepository) ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]db.TaskComment, *apperror.AppError) {
	comments, err := r.q.ListTaskComments(ctx, taskID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list task comments").WithDetail("error", err.Error())
	}

	return comments, nil
}
