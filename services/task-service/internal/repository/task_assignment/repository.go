package taskassigment

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)


type taskAssignmentRepository struct {
	q db.Querier
}

func NewTaskAssignmentRepository (q db.Querier) TaskAssignmentRepository {
	return &taskAssignmentRepository{q: q}
}


func(r *taskAssignmentRepository) AssignTask(ctx context.Context, params db.AssignTaskParams)(*db.TaskAssignment, *apperror.AppError) {
	assignee, err := r.q.AssignTask(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to assignee task").WithDetail("error", err.Error())		
	}
	
	return &assignee, nil
}

func(r *taskAssignmentRepository) GetActiveTaskAssignment(ctx context.Context, params db.GetActiveTaskAssignmentParams)(*db.TaskAssignment, *apperror.AppError) {
	assignee, err := r.q.GetActiveTaskAssignment(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to get active task assignment").WithDetail("error", err.Error())		
	}
	
	return &assignee, nil
}

func(r *taskAssignmentRepository) UnassignTask(ctx context.Context, params db.UnassignTaskParams)(*db.TaskAssignment, *apperror.AppError) {
	assignee, err := r.q.UnassignTask(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to unassignee task").WithDetail("error", err.Error())		
	}
	
	return &assignee, nil
}

func(r *taskAssignmentRepository) ListTaskAssignments(ctx context.Context, taskID uuid.UUID)([]db.TaskAssignment, *apperror.AppError) {
	assignee, err := r.q.ListTaskAssignments(ctx, taskID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list task assignments").WithDetail("error", err.Error())		
	}
	
	return assignee, nil
}

func(r *taskAssignmentRepository) ListActiveAssignmentsByAssignee(ctx context.Context, params db.ListActiveAssignmentsByAssigneeParams)([]db.TaskAssignment, *apperror.AppError) {
	assignee, err := r.q.ListActiveAssignmentsByAssignee(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to list active assignments by assignee").WithDetail("error", err.Error())		
	}
	
	return assignee, nil
}