package task

import (
	"context"
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/task_service/models/v1"
	taskv1 "github.com/rijum8906/relay/packages/pb/task_service/task/v1"
	"github.com/rijum8906/relay/services/task-service/internal/authz"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

func (s *service) addProjectMember(ctx context.Context, req *taskv1.AddProjectMemberRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
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

	currentMembershipRow, err := s.q.GetActiveProjectMembership(ctx, db.GetActiveProjectMembershipParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if _, appErr = utils.QueryOne(currentMembershipRow, err, "project membership not found", "failed to get project membership"); appErr == nil {
		return nil, apperror.ErrConflict.WithMessage("project member already exists")
	} else if appErr.Code != apperror.CodeNotFound {
		return nil, appErr
	}

	membershipRow, err := s.q.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
	membership, appErr := utils.QueryOne(membershipRow, err, "", "failed to add project member")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.ProjectRoleTuple(projectID, role, userID),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapProjectMembership(membership), nil
}

func (s *service) removeProjectMember(ctx context.Context, req *taskv1.RemoveProjectMemberRequest, userInfo *dto.UserInfo) (*corev1.SuccessResponse, *apperror.AppError) {
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

	membershipRow, err := s.q.RemoveProjectMember(ctx, db.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	membership, appErr := utils.QueryOne(membershipRow, err, "project membership not found", "failed to remove project member")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil {
		if appErr := s.tuples.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			authz.DeleteTuple(authz.ProjectRoleTuple(projectID, membership.Role, userID)),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *service) updateProjectMemberRole(ctx context.Context, req *taskv1.UpdateProjectMemberRoleRequest, userInfo *dto.UserInfo) (*modelsv1.ProjectMembership, *apperror.AppError) {
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

	var currentMembership *db.ProjectMembership
	if s.tuples != nil {
		currentMembershipRow, err := s.q.GetActiveProjectMembership(ctx, db.GetActiveProjectMembershipParams{
			ProjectID: projectID,
			UserID:    userID,
		})
		currentMembership, appErr = utils.QueryOne(currentMembershipRow, err, "project membership not found", "failed to get project membership")
		if appErr != nil {
			return nil, appErr
		}
	}

	membershipRow, err := s.q.UpdateProjectMemberRole(ctx, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
	membership, appErr := utils.QueryOne(membershipRow, err, "project membership not found", "failed to update project member role")
	if appErr != nil {
		return nil, appErr
	}

	if s.tuples != nil && currentMembership.Role != membership.Role {
		if appErr := s.tuples.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			authz.DeleteTuple(authz.ProjectRoleTuple(projectID, currentMembership.Role, userID)),
		}); appErr != nil {
			return nil, appErr
		}
		if appErr := s.tuples.Write(ctx, []client.ClientTupleKey{
			authz.ProjectRoleTuple(projectID, membership.Role, userID),
		}); appErr != nil {
			return nil, appErr
		}
	}

	return mapProjectMembership(membership), nil
}

func (s *service) listProjectMembers(ctx context.Context, req *taskv1.ListProjectMembersRequest, userInfo *dto.UserInfo) (*taskv1.ListProjectMembersResponse, *apperror.AppError) {
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

	memberRows, err := s.q.ListProjectMembers(ctx, projectID)
	members, appErr := utils.QueryMany(memberRows, err, "failed to list project members")
	if appErr != nil {
		return nil, appErr
	}

	return &taskv1.ListProjectMembersResponse{
		Members: mapProjectMemberships(members),
	}, nil
}

func normalizeRole(value string) (string, *apperror.AppError) {
	role := strings.TrimSpace(value)
	if role == "" {
		role = string(taskpermissions.RoleMember)
	}

	switch role {
	case string(taskpermissions.RoleOwner), string(taskpermissions.RoleAdmin), string(taskpermissions.RoleMember):
		return role, nil
	default:
		return "", apperror.ErrValidation.WithMessage("invalid project member role").WithDetail("field", "role")
	}
}
