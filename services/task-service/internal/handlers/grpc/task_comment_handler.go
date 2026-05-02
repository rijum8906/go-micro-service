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

func (h *TaskHandler) CreateTaskComment(ctx context.Context, req *taskv1.CreateTaskCommentRequest) (*modelsv1.TaskComment, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task comment request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("create task comment user metadata is required"))
	}

	result, appErr := h.taskCommentService.CreateTaskComment(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateTaskComment(ctx context.Context, req *taskv1.UpdateTaskCommentRequest) (*modelsv1.TaskComment, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update task comment request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("update task comment user metadata is required"))
	}

	result, appErr := h.taskCommentService.UpdateTaskComment(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) DeleteTaskComment(ctx context.Context, req *taskv1.DeleteTaskCommentRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("delete task comment request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("delete task comment user metadata is required"))
	}

	result, appErr := h.taskCommentService.DeleteTaskComment(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTaskComments(ctx context.Context, req *taskv1.ListTaskCommentsRequest) (*taskv1.ListTaskCommentsResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list task comments request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list task comments user metadata is required"))
	}

	result, appErr := h.taskCommentService.ListTaskComments(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
