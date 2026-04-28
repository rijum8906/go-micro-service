// Package grpc contains task-service gRPC handler placeholders until the API is defined.
package grpc

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (h *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("create task user metadata is required"))
	}

	result, appErr := h.taskService.CreateTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}

	result, appErr := h.taskService.GetTask(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}

	result, appErr := h.taskService.ListTasksByProject(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
