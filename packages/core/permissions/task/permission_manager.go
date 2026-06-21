package permissions

import (
	"context"
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

type PermissionManager struct {
	tupleManager coreopenfga.TuppleManager
}

func NewPermissionManager(fgaClient *coreopenfga.Client) *PermissionManager {
	return NewPermissionManagerWithTupleManager(coreopenfga.NewTupleManager(fgaClient))
}

func NewPermissionManagerWithTupleManager(tupleManager coreopenfga.TuppleManager) *PermissionManager {
	return &PermissionManager{tupleManager: tupleManager}
}

func (pm *PermissionManager) CreateCustomRole(ctx context.Context, userID, projectID, role string, grants ...PermissionGrant) *apperror.AppError {
	if appErr := validateCustomRoleInput(userID, projectID, role, grants...); appErr != nil {
		return appErr
	}

	roleObj := GenerateCustomRoleObject(projectID, role)
	tuples := []client.ClientTupleKey{
		{User: "project:" + projectID, Relation: ResourceProject, Object: roleObj},
		{User: "user:" + userID, Relation: "assignee", Object: roleObj},
	}

	for _, grant := range grants {
		tuples = append(tuples, client.ClientTupleKey{
			User:     roleObj,
			Relation: "granted_to",
			Object:   GeneratePermissionObject(projectID, grant.Resource, grant.Permission),
		})
	}

	return pm.tupleManager.Write(ctx, tuples)
}

func (pm *PermissionManager) DeleteCustomRole(ctx context.Context, userID, projectID, role string, grants ...PermissionGrant) *apperror.AppError {
	if appErr := validateCustomRoleInput(userID, projectID, role, grants...); appErr != nil {
		return appErr
	}

	roleObj := GenerateCustomRoleObject(projectID, role)
	tuples := []client.ClientTupleKeyWithoutCondition{
		{User: "project:" + projectID, Relation: ResourceProject, Object: roleObj},
		{User: "user:" + userID, Relation: "assignee", Object: roleObj},
	}

	for _, grant := range grants {
		tuples = append(tuples, client.ClientTupleKeyWithoutCondition{
			User:     roleObj,
			Relation: "granted_to",
			Object:   GeneratePermissionObject(projectID, grant.Resource, grant.Permission),
		})
	}

	return pm.tupleManager.Delete(ctx, tuples)
}

func validateCustomRoleInput(userID, projectID, role string, grants ...PermissionGrant) *apperror.AppError {
	if strings.TrimSpace(userID) == "" {
		return apperror.ErrValidation.WithMessage("user id is required").WithDetail("field", "user_id")
	}
	if strings.TrimSpace(projectID) == "" {
		return apperror.ErrValidation.WithMessage("project id is required").WithDetail("field", "project_id")
	}
	if !IsValidCustomRole(role) {
		return apperror.ErrValidation.WithMessage("invalid custom role").WithDetail("field", "role")
	}
	for _, grant := range grants {
		if !IsValidResourcePermission(grant.Resource, grant.Permission) {
			return apperror.ErrValidation.WithMessage("invalid permission grant").WithDetail("permission", grant.Permission)
		}
	}
	return nil
}
