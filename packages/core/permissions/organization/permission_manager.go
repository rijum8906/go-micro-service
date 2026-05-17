package permissions

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
)

type PermissionManager struct {
	tupleManager coreopenfga.TuppleManager
}

func NewPermissionManager(fgaClient *coreopenfga.Client) *PermissionManager {
	return &PermissionManager{
		tupleManager: coreopenfga.NewTupleManager(fgaClient),
	}
}

func (pm *PermissionManager) CreateOrgRole(ctx context.Context, userID, orgID, permission string) *apperror.AppError {
	return pm.tupleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userID,
			Relation: permission,
			Object:   "organization:" + orgID,
		},
	})
}

func (pm *PermissionManager) DeleteOrgRole(ctx context.Context, userID, orgID, permission string) *apperror.AppError {
	return pm.tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:" + userID,
			Relation: permission,
			Object:   "organization:" + orgID,
		},
	})
}

func (pm *PermissionManager) CreateOrgTeamRole(ctx context.Context, userID, teamID, permission string) *apperror.AppError {
	return pm.tupleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userID,
			Relation: permission,
			Object:   "team:" + teamID,
		},
	})
}

func (pm *PermissionManager) DeleteOrgTeamRole(ctx context.Context, userID, teamID, permission string) *apperror.AppError {
	return pm.tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:" + userID,
			Relation: permission,
			Object:   "team:" + teamID,
		},
	})
}

func (pm *PermissionManager) CreateCustomOrgRole(ctx context.Context, userID, orgID, role string, permissions ...string) *apperror.AppError {
	roleObj := GenerateCustomRoleObject(orgID, role)
	tuples := []client.ClientTupleKey{
		{
			User:     "organization:" + orgID,
			Relation: "organization",
			Object:   roleObj,
		},
		{
			User:     "user:" + userID,
			Relation: "assignee",
			Object:   roleObj,
		},
	}

	for _, permission := range permissions {
		tuples = append(tuples, client.ClientTupleKey{
			User:     roleObj,
			Relation: "granted_to",
			Object:   GeneratePermissionObject(orgID, permission),
		})
	}

	return pm.tupleManager.Write(ctx, tuples)
}

func (pm *PermissionManager) DeleteCustomOrgRole(ctx context.Context, userID, orgID, role string) *apperror.AppError {
	roleObj := GenerateCustomRoleObject(orgID, role)
	return pm.tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "organization:" + orgID,
			Relation: "organization",
			Object:   roleObj,
		},
	})
}

func (pm *PermissionManager) CreateCustomOrgTeamRole(ctx context.Context, userID, teamID, role string, permissions ...string) *apperror.AppError {
	roleObj := GenerateCustomRoleObject(teamID, role)
	tuples := []client.ClientTupleKey{
		{
			User:     "team:" + teamID,
			Relation: "team",
			Object:   roleObj,
		},
		{
			User:     "user:" + userID,
			Relation: "assignee",
			Object:   roleObj,
		},
	}

	for _, permission := range permissions {
		tuples = append(tuples, client.ClientTupleKey{
			User:     roleObj,
			Relation: "granted_to",
			Object:   GeneratePermissionObject(teamID, permission),
		})
	}

	return pm.tupleManager.Write(ctx, tuples)
}

func (pm *PermissionManager) DeleteCustomOrgTeamRole(ctx context.Context, userID, teamID, role string) *apperror.AppError {
	roleObj := GenerateCustomRoleObject(teamID, role)
	return pm.tupleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "team:" + teamID,
			Relation: "team",
			Object:   roleObj,
		},
	})
}
