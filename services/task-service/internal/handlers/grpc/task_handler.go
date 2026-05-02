// Package grpc contains task-service gRPC handler placeholders until the API is defined.
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

func (h *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("create task user metadata is required"))
	}

	result, appErr := h.taskService.CreateTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("get task user metadata is required"))
	}

	result, appErr := h.taskService.GetTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTasksByProject(ctx context.Context, req *taskv1.ListTasksByProjectRequest) (*taskv1.ListTasksByProjectResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list tasks by project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list tasks by project user metadata is required"))
	}

	result, appErr := h.taskService.ListTasksByProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("update task user metadata is required"))
	}

	result, appErr := h.taskService.UpdateTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("delete task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("delete task user metadata is required"))
	}

	result, appErr := h.taskService.DeleteTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ArchiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("archive task request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("archive task user metadata is required"))
	}

	result, appErr := h.taskService.ArchiveTask(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateTaskStatus(ctx context.Context, req *taskv1.UpdateTaskStatusRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update task status request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("update task status user metadata is required"))
	}

	result, appErr := h.taskService.UpdateTaskStatus(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateTaskProgress(ctx context.Context, req *taskv1.UpdateTaskProgressRequest) (*modelsv1.Task, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update task progress request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("update task progress user metadata is required"))
	}

	result, appErr := h.taskService.UpdateTaskProgress(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTasksByOrganization(ctx context.Context, req *taskv1.ListTasksByOrganizationRequest) (*taskv1.ListTasksByOrganizationResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list tasks by organization request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list tasks by organization user metadata is required"))
	}

	result, appErr := h.taskService.ListTasksByOrganization(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTasksByParent(ctx context.Context, req *taskv1.ListTasksByParentRequest) (*taskv1.ListTasksByParentResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list tasks by parent request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list tasks by parent user metadata is required"))
	}

	result, appErr := h.taskService.ListTasksByParent(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListTasksByCreator(ctx context.Context, req *taskv1.ListTasksByCreatorRequest) (*taskv1.ListTasksByCreatorResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list tasks by creator request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list tasks by creator user metadata is required"))
	}

	result, appErr := h.taskService.ListTasksByCreator(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
