package authz

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	coredto "github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/services/task-service/internal/db"
	projectmembershiprepo "github.com/rijum8906/relay/services/task-service/internal/repository/project_membership"
	taskrepo "github.com/rijum8906/relay/services/task-service/internal/repository/task"
	"github.com/rijum8906/relay/services/task-service/internal/utils"
)

type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
	RoleOwner  Role = "owner"
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
	memberships projectmembershiprepo.ProjectMembershipRepository
	tasks       taskrepo.TaskRepository
}

func NewAuthorizer(memberships projectmembershiprepo.ProjectMembershipRepository, tasks taskrepo.TaskRepository) (Authorizer, *apperror.AppError) {
	if memberships == nil || tasks == nil {
		return nil, apperror.ErrInternal.WithMessage("failed to initialize authorizer")
	}

	return &authorizer{
		memberships: memberships,
		tasks:       tasks,
	}, nil
}

func (a *authorizer) RequireProjectRole(ctx context.Context, projectID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.ProjectMembership, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	memberships, appErr := a.memberships.GetActiveProjectMembership(ctx, db.GetActiveProjectMembershipParams{
		ProjectID: projectID,
		UserID:    userID,
	})

	if appErr != nil {
		if appErr.Code == apperror.CodeNotFound {
			return nil, apperror.ErrForbidden.WithMessage("you do not have access to this project")
		}
		return nil, appErr
	}

	if !hasMinRole(memberships.Role, minRole) {
		return nil, apperror.ErrForbidden.WithMessage("insufficient project role")
	}

	return memberships, nil
}

func (a *authorizer) RequireTaskRole(ctx context.Context, taskID uuid.UUID, userInfo *coredto.UserInfo, minRole Role) (*db.Task, *apperror.AppError) {
	userID, appErr := utils.ValidateUserInfo(userInfo)
	if appErr != nil {
		return nil, appErr
	}

	task, appErr := a.tasks.GetTask(ctx, taskID)
	if appErr != nil {
		return nil, appErr
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
