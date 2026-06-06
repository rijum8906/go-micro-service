// Package grpc contains task-service gRPC handler placeholders until the API is defined.
package grpc

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	taskservice "github.com/rijum8906/relay/services/task-service/internal/services/task"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	service taskservice.TaskService
}

func NewTaskHandler(service taskservice.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}

	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("create task user metadata is required"))
	}

	result, appErr := h.service.CreateTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}
	result, appErr := h.service.GetTask(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}
	return result, nil
}

func (h *TaskHandler) ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}
	result, appErr := h.service.ListTasksByProject(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}
	return result, nil
}
