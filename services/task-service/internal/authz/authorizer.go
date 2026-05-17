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

type Role = taskpermissions.Role

const (
	RoleMember Role = taskpermissions.RoleMember
	RoleAdmin  Role = taskpermissions.RoleAdmin
	RoleOwner  Role = taskpermissions.RoleOwner
)

var roleRank = map[Role]int{
	RoleMember: 1,
	RoleAdmin:  2,
	RoleOwner:  3,
}

type Authorizer interface {
	RequireProjectRole(ctx context.Context, projectID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.ProjectMembership, *apperror.AppError)
	RequireTaskRole(ctx context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.Task, *apperror.AppError)
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

func (a *authorizer) RequireProjectRole(ctx context.Context, projectID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.ProjectMembership, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	if a.tuples != nil {
		if appErr = a.requireFGA(ctx, userID, projectRelation(minRole), fgaObject("project", projectID)); appErr != nil {
			return nil, appErr
		}
		return &db.ProjectMembership{
			ProjectID: projectID,
			UserID:    userID,
			Role:      string(minRole),
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

	if !hasMinRole(membership.Role, minRole) {
		return nil, apperror.ErrForbidden.WithMessage("insufficient project role")
	}

	return membership, nil
}

func (a *authorizer) RequireTaskRole(ctx context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.Task, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	taskRow, err := a.q.GetTask(ctx, taskID)
	task, appErr := utils.QueryOne(taskRow, err, "task not found", "failed to get task")
	if appErr != nil {
		return nil, appErr
	}

	if a.tuples != nil {
		if appErr = a.requireFGA(ctx, userID, taskRelation(minRole), fgaObject("task", taskID)); appErr != nil {
			return nil, appErr
		}
		return task, nil
	}

	if task.ProjectID.Valid {
		if _, appErr := a.RequireProjectRole(ctx, task.ProjectID.Bytes, userInfo, minRole); appErr != nil {
			return nil, appErr
		}
		return task, nil
	}

	if task.CreatedBy != userID {
		return nil, apperror.ErrForbidden.WithMessage("you do not have access to this task")
	}

	return task, nil
}

func hasMinRole(actual string, required Role) bool {
	actualRole := Role(actual)
	return roleRank[actualRole] >= roleRank[required]
}

func (a *authorizer) requireFGA(ctx context.Context, userID uuid.UUID, relation, object string) *apperror.AppError {
	res, appErr := a.tuples.Check(ctx, client.ClientCheckRequest{
		User:     fgaObject("user", userID),
		Relation: relation,
		Object:   object,
	})
	if appErr != nil {
		return appErr
	}
	if res == nil || !res.GetAllowed() {
		return apperror.ErrForbidden.WithMessage("permission denied")
	}

	return nil
}

func projectRelation(role Role) string {
	switch role {
	case RoleOwner:
		return taskpermissions.PermissionCanDelete
	case RoleAdmin:
		return taskpermissions.PermissionCanManageTasks
	default:
		return taskpermissions.PermissionCanView
	}
}

func taskRelation(role Role) string {
	switch role {
	case RoleOwner:
		return taskpermissions.PermissionCanDelete
	case RoleAdmin:
		return taskpermissions.PermissionCanManage
	default:
		return taskpermissions.PermissionCanView
	}
}

func fgaObject(objectType string, id uuid.UUID) string {
	return fmt.Sprintf("%s:%s", objectType, id)
}
