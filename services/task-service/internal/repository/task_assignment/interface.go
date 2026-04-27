package taskassigment

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/task-service/internal/db"
)



type TaskAssignmentRepository interface {
	AssignTask(ctx context.Context, params db.AssignTaskParams)(*db.TaskAssignment, *apperror.AppError)
	GetActiveTaskAssignment(ctx context.Context, params db.GetActiveTaskAssignmentParams) (*db.TaskAssignment, *apperror.AppError)
	UnassignTask(ctx context.Context, params db.UnassignTaskParams)(*db.TaskAssignment, *apperror.AppError)
	ListTaskAssignments(ctx context.Context, taskID uuid.UUID) ([]db.TaskAssignment, *apperror.AppError)
	ListActiveAssignmentsByAssignee(ctx context.Context, params db.ListActiveAssignmentsByAssigneeParams)([]db.TaskAssignment, *apperror.AppError)
	
}

