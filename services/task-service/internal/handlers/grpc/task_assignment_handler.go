package grpc

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (h *TaskHandler) AssignTask(ctx context.Context, req *taskv1.AssignTaskRequest) (*modelsv1.TaskAssignment, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("assign task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("assign task user metadata is required"))
	}

	result, appErr := h.taskAssignmentService.AssignTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UnassignTask(ctx context.Context, req *taskv1.UnassignTaskRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("unassign task request is required"))
	}

	result, appErr := h.taskAssignmentService.UnassignTask(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ReassignTask(ctx context.Context, req *taskv1.ReassignTaskRequest) (*modelsv1.TaskAssignment, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("reassign task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("reassign task user metadata is required"))
	}

	result, appErr := h.taskAssignmentService.ReassignTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTaskAssignments(ctx context.Context, req *taskv1.ListTaskAssignmentsRequest) (*taskv1.ListTaskAssignmentsResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list task assignments request is required"))
	}

	result, appErr := h.taskAssignmentService.ListTaskAssignments(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
