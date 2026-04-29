package grpc

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (h *TaskHandler) AddProjectMember(ctx context.Context, req *taskv1.AddProjectMemberRequest) (*modelsv1.ProjectMembership, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("add project member request is required"))
	}

	result, appErr := h.projectMembershipService.AddProjectMember(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) RemoveProjectMember(ctx context.Context, req *taskv1.RemoveProjectMemberRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("remove project member request is required"))
	}

	result, appErr := h.projectMembershipService.RemoveProjectMember(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) UpdateProjectMemberRole(ctx context.Context, req *taskv1.UpdateProjectMemberRoleRequest) (*modelsv1.ProjectMembership, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("update project member role request is required"))
	}

	result, appErr := h.projectMembershipService.UpdateProjectMemberRole(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *TaskHandler) ListProjectMembers(ctx context.Context, req *taskv1.ListProjectMembersRequest) (*taskv1.ListProjectMembersResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("list project members request is required"))
	}

	result, appErr := h.projectMembershipService.ListProjectMembers(ctx, req)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
