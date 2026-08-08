package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	taskpermissions "github.com/rijum8906/relay/packages/core/permissions/task"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

type Authorizer interface {
	RequireProjectPermission(ctx context.Context, projectID uuid.UUID, userInfo *coredto.UserInfo, permission string) (*db.ProjectMembership, *apperror.AppError)
	RequireTaskPermission(ctx context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, permission string) (*db.Task, *apperror.AppError)
}

type authorizer struct {
	q      db.Querier
	tuples coreopenfga.TuppleManager
}

func NewAuthorizer(q db.Querier, tuples coreopenfga.TuppleManager) (Authorizer, *apperror.AppError) {
	if q == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize authorizer").WithDetail("queries", "task queries must be configured")
	}

	return &authorizer{
		q:      q,
		tuples: tuples,
	}, nil
}

func (a *authorizer) RequireProjectPermission(ctx context.Context, projectID uuid.UUID, userInfo *coredto.UserInfo, permission string) (*db.ProjectMembership, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}
	if !taskpermissions.IsProjectPermission(permission) {
		return nil, apperror.ErrValidation.WithMessage("invalid project permission").WithDetail("permission", permission)
	}

	if a.tuples != nil {
		allowed, appErr := a.checkFGA(ctx, userID, permission, fgaObject("project", projectID))
		if appErr != nil {
			return nil, appErr
		}
		if !allowed {
			allowed, appErr = a.checkFGA(ctx, userID, "allowed", taskpermissions.GeneratePermissionObject(projectID.String(), taskpermissions.ResourceProject, permission))
			if appErr != nil {
				return nil, appErr
			}
		}
		if !allowed {
			return nil, apperror.ErrForbidden.WithMessage("permission denied")
		}

		return &db.ProjectMembership{
			ProjectID: projectID,
			UserID:    userID,
		}, nil
	}

	membershipRow, err := a.q.GetActiveProjectMembership(ctx, db.GetActiveProjectMembershipParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	membership, appErr := utils.QueryOne(membershipRow, err, "project membership not found", "failed to get project membership")
	if appErr != nil {
		if appErr.Code == apperror.CodeNotFound {
			return nil, apperror.ErrForbidden.WithMessage("you do not have access to this project")
		}
		return nil, appErr
	}

	if !hasDefaultRolePermission(membership.Role, permission) {
		return nil, apperror.ErrForbidden.WithMessage("insufficient project permission")
	}

	return membership, nil
}

func (a *authorizer) RequireTaskPermission(ctx context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, permission string) (*db.Task, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}
	if !taskpermissions.IsTaskPermission(permission) {
		return nil, apperror.ErrValidation.WithMessage("invalid task permission").WithDetail("permission", permission)
	}

	taskRow, err := a.q.GetTask(ctx, taskID)
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to get task")
	if appErr != nil {
		return nil, appErr
	}

	if a.tuples != nil {
		allowed, appErr := a.checkFGA(ctx, userID, permission, fgaObject("task", taskID))
		if appErr != nil {
			return nil, appErr
		}
		if !allowed && task.ProjectID.Valid {
			allowed, appErr = a.checkFGA(ctx, userID, "allowed", taskpermissions.GeneratePermissionObject(uuid.UUID(task.ProjectID.Bytes).String(), taskpermissions.ResourceTask, permission))
			if appErr != nil {
				return nil, appErr
			}
		}
		if !allowed {
			return nil, apperror.ErrForbidden.WithMessage("permission denied")
		}

		return task, nil
	}

	if task.ProjectID.Valid {
		projectPermission := projectPermissionForTask(permission)
		if _, appErr := a.RequireProjectPermission(ctx, task.ProjectID.Bytes, userInfo, projectPermission); appErr != nil {
			return nil, appErr
		}
		return task, nil
	}

	if task.CreatedBy != userID {
		return nil, apperror.ErrForbidden.WithMessage("you do not have access to this task")
	}

	return task, nil
}

func (a *authorizer) checkFGA(ctx context.Context, userID uuid.UUID, relation, object string) (bool, *apperror.AppError) {
	res, appErr := a.tuples.Check(ctx, client.ClientCheckRequest{
		User:     fgaObject("user", userID),
		Relation: relation,
		Object:   object,
	})
	if appErr != nil {
		return false, appErr
	}

	return res != nil && res.GetAllowed(), nil
}

func hasDefaultRolePermission(role, permission string) bool {
	for _, defaultPermission := range taskpermissions.GetDefaultPermissionsForRole(role) {
		if defaultPermission == permission {
			return true
		}
	}
	return false
}

func projectPermissionForTask(permission string) string {
	switch permission {
	case taskpermissions.PermissionCanView:
		return taskpermissions.PermissionCanView
	case taskpermissions.PermissionCanEdit,
		taskpermissions.PermissionCanUpdateStatus,
		taskpermissions.PermissionCanUpdateProgress,
		taskpermissions.PermissionCanComment:
		return taskpermissions.PermissionCanContributeTasks
	case taskpermissions.PermissionCanManage,
		taskpermissions.PermissionCanAssign,
		taskpermissions.PermissionCanArchive,
		taskpermissions.PermissionCanDelete:
		return taskpermissions.PermissionCanManageTasks
	default:
		return taskpermissions.PermissionCanView
	}
}

func fgaObject(objectType string, id uuid.UUID) string {
	return fmt.Sprintf("%s:%s", objectType, id)
}
