package taskassigment

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	taskassignmentrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task_assignment"
)

type TaskAssignmentService interface {
	AssignTask(ctx context.Context, req *taskv1.AssignTaskRequest, userInfo *dto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError)
	UnassignTask(ctx context.Context, req *taskv1.UnassignTaskRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError)
	ReassignTask(ctx context.Context, req *taskv1.ReassignTaskRequest, userInfo *dto.UserInfo) (*modelsv1.TaskAssignment, *apperror.AppError)
	ListTaskAssignments(ctx context.Context, req *taskv1.ListTaskAssignmentsRequest, userInfo *dto.UserInfo) (*taskv1.ListTaskAssignmentsResponse, *apperror.AppError)
}

type service struct {
	repo  taskassignmentrepo.TaskAssignmentRepository
	authz authz.Authorizer
}

func NewTaskAssignmentService(repo taskassignmentrepo.TaskAssignmentRepository, authz authz.Authorizer) (TaskAssignmentService, *apperror.AppError) {
	if repo == nil || authz == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize task assignment service").WithDetail("repo", "task assignment repository must be configured")
	}

	return &service{
		repo:  repo,
		authz: authz,
	}, nil
}
