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

func (h *TaskHandler) CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest) (*modelsv1.Project, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("create project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("create project user metadata is required"))
	}

	result, appErr := h.projectService.CreateProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) GetProject(ctx context.Context, req *taskv1.GetProjectRequest) (*modelsv1.Project, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("get project user metadata is required"))
	}

	result, appErr := h.projectService.GetProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest) (*modelsv1.Project, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("update project user metadata is required"))
	}

	result, appErr := h.projectService.UpdateProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) CompleteProject(ctx context.Context, req *taskv1.CompleteProjectRequest) (*modelsv1.Project, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("complete project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("complete project user metadata is required"))
	}

	result, appErr := h.projectService.CompleteProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ArchiveProject(ctx context.Context, req *taskv1.ArchiveProjectRequest) (*modelsv1.Project, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("archive project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("archive project user metadata is required"))
	}

	result, appErr := h.projectService.ArchiveProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("delete project request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("delete project user metadata is required"))
	}

	result, appErr := h.projectService.DeleteProject(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListProjects(ctx context.Context, req *taskv1.ListProjectsRequest) (*taskv1.ListProjectsResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list projects request is required"))
	}

	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrUnAuthenticated.WithMessage("list projects user metadata is required"))
	}

	result, appErr := h.projectService.ListProjects(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
