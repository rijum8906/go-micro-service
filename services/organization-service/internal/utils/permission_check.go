package utils

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func CheckCanChangeMembershipStatus(
	ctx context.Context,
	tuppleManager coreopenfga.TuppleManager,
	actorMembership, targetMembership *db.OrganizationMembership,
) *apperror.AppError {
	checkReq := client.ClientCheckRequest{}
	if permissions.IsValidRole(actorMembership.Role) {
		checkReq = client.ClientCheckRequest{
			User:     "user:" + actorMembership.UserID.String(),
			Relation: actorMembership.Role,
			Object:   "organization:" + actorMembership.OrganizationID.String(),
		}
	} else {
		checkReq = client.ClientCheckRequest{
			User:     "user:" + actorMembership.UserID.String(),
			Relation: "allowed",
			Object:   permissions.GeneratePermissionObject(actorMembership.OrganizationID.String(), permissions.PermissionCanChangeMemberStatus),
		}
	}
	res, appErr := tuppleManager.Check(ctx, checkReq)
	if appErr != nil {
		return appErr
	}
	if !*res.Allowed {
		return apperror.ErrPermissionDenied.WithMessage("you do not have permission to change this membership's status")
	}
	return nil
}
