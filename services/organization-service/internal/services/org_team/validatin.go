package orgteam

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
)

func parseCreateOrgTeamReq(req *org_teamv1.CreateOrgTeamRequest) (orgID uuid.UUID, membershipID uuid.UUID, appErr *apperror.AppError) {
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid organization id")
	}
	membershipID, err = uuid.Parse(req.MembershipId)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid membership id")
	}

	if req.Name == "" {
		return uuid.UUID{}, uuid.UUID{}, apperror.ErrValidation.WithMessage("name cannot be empty")
	}

	if req.Description == "" {
		return uuid.UUID{}, uuid.UUID{}, apperror.ErrValidation.WithMessage("description cannot be empty")
	}

	return orgID, membershipID, nil
}
