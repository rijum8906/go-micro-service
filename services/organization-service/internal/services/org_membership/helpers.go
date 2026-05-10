package orgmembership

import (
	"context"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

// removeRole removes the role from the user
// whether the role is a standard role or custom role
func (s *orgMembershipService) removeRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	if permissions.IsValidRole(targetMembership.Role) {
		s.tuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: targetMembership.Role,
				Object:   "organization:" + targetMembership.OrganizationID.String(),
			},
		})
	} else {
		s.tuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: "allowed",
				Object:   permissions.GenerateCustomRoleObject(targetMembership.OrganizationID.String(), targetMembership.Role),
			},
		})
	}
	return nil
}

func (s *orgMembershipService) addRole(ctx context.Context, targetMembership *db.OrganizationMembership) *apperror.AppError {
	if permissions.IsValidRole(targetMembership.Role) {
		s.tuppleManager.Write(ctx, []client.ClientTupleKey{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: targetMembership.Role,
				Object:   "organization:" + targetMembership.OrganizationID.String(),
			},
		})
	} else {
		s.tuppleManager.Write(ctx, []client.ClientTupleKey{
			{
				User:     "user:" + targetMembership.UserID.String(),
				Relation: "allowed",
				Object:   permissions.GenerateCustomRoleObject(targetMembership.OrganizationID.String(), targetMembership.Role),
			},
		})
	}
	return nil
}

type membershipData struct {
	target *db.OrganizationMembership
	actor  *db.OrganizationMembership
}

func (s *orgMembershipService) retrieveMemberships(ctx context.Context, membershipID uuid.UUID, userID string) (*membershipData, *apperror.AppError) {
	target, err := s.q.GetOrganizationMembership(ctx, membershipID)
	if err != nil {
		return nil, coreutils.ParseDBError(err, "target membership")
	}

	actor, err := s.q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
		OrganizationID: target.OrganizationID,
		UserID:         uuid.MustParse(userID),
	})
	if err != nil {
		return nil, coreutils.ParseDBError(err, "actor membership")
	}

	return &membershipData{
		target: &target,
		actor:  &actor,
	}, nil
}
