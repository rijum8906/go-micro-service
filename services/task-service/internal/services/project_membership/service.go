package projectmembership

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) AddProjectMember(ctx context.Context, req *taskv1.AddProjectMemberRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("add project member request is required")
	}

	_, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	projectID, appErr := requiredUUID(req.GetProjectId(), "project_id", "project id is required")
	if appErr != nil {
		return nil, appErr
	}

	userID, appErr := requiredUUID(req.GetUserId(), "user_id", "user id is required")
	if appErr != nil {
		return nil, appErr
	}

	role, appErr := normalizeRole(req.GetRole())
	if appErr != nil {
		return nil, appErr
	}
	requiredRole := authz.RoleAdmin
	if role == string(authz.RoleOwner) {
		requiredRole = authz.RoleOwner
	}
	if _, appErr = s.authz.RequireProjectRole(ctx, projectID, userInfo, requiredRole); appErr != nil {
		return nil, appErr
	}

	if _, appErr = s.repo.GetActiveProjectMembership(ctx, db.GetActiveProjectMembershipParams{
		ProjectID: projectID,
		UserID:    userID,
	}); appErr == nil {
		return nil, apperror.ErrConflict.WithMessage("project member already exists")
	} else if appErr.Code != apperror.CodeNotFound {
		return nil, appErr
	}

	membership, appErr := s.repo.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapProjectMembership(membership), nil
}

func (s *service) RemoveProjectMember(ctx context.Context, req *taskv1.RemoveProjectMemberRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("remove project member request is required")
	}

	userIDCheck, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userIDCheck

	projectID, appErr := requiredUUID(req.GetProjectId(), "project_id", "project id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireProjectRole(ctx, projectID, userInfo, authz.RoleAdmin); appErr != nil {
		return nil, appErr
	}

	userID, appErr := requiredUUID(req.GetUserId(), "user_id", "user id is required")
	if appErr != nil {
		return nil, appErr
	}

	if _, appErr = s.repo.RemoveProjectMember(ctx, db.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) UpdateProjectMemberRole(ctx context.Context, req *taskv1.UpdateProjectMemberRoleRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update project member role request is required")
	}

	userIDCheck, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userIDCheck

	projectID, appErr := requiredUUID(req.GetProjectId(), "project_id", "project id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireProjectRole(ctx, projectID, userInfo, authz.RoleOwner); appErr != nil {
		return nil, appErr
	}

	userID, appErr := requiredUUID(req.GetUserId(), "user_id", "user id is required")
	if appErr != nil {
		return nil, appErr
	}

	role, appErr := normalizeRole(req.GetRole())
	if appErr != nil {
		return nil, appErr
	}

	membership, appErr := s.repo.UpdateProjectMemberRole(ctx, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
	if appErr != nil {
		return nil, appErr
	}

	return mapProjectMembership(membership), nil
}

func (s *service) ListProjectMembers(ctx context.Context, req *taskv1.ListProjectMembersRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectMembersResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("list project members request is required")
	}

	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	_ = userID

	projectID, appErr := requiredUUID(req.GetProjectId(), "project_id", "project id is required")
	if appErr != nil {
		return nil, appErr
	}
	if _, appErr = s.authz.RequireProjectRole(ctx, projectID, userInfo, authz.RoleMember); appErr != nil {
		return nil, appErr
	}

	members, appErr := s.repo.ListProjectMembers(ctx, projectID)
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListProjectMembersResponse{
		Members: mapProjectMemberships(members),
	}, nil
}

func requiredUUID(value, field, requiredMessage string) (uuid.UUID, *apperror.AppError) {
	if strings.TrimSpace(value) == "" {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage(requiredMessage)
	}

	id, appErr := utils.NewUUID(value)
	if appErr != nil {
		return uuid.UUID{}, appErr.WithDetail("field", field)
	}

	return id, nil
}

func normalizeRole(value string) (string, *apperror.AppError) {
	role := strings.TrimSpace(value)
	if role == "" {
		role = "member"
	}

	switch role {
	case "owner", "admin", "member":
		return role, nil
	default:
		return "", apperror.ErrValidation.WithMessage("invalid project member role").WithDetail("field", "role")
	}
}
