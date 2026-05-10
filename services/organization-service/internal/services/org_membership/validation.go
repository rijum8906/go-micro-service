package orgmembership

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
)

func validateChangeOrganizationStatusReq(req *org_membershipv1.ChangeOrgMembershipStatusReq) error {
	if err := uuid.Validate(req.OrganizationMembershipId); err != nil {
		return apperror.ErrValidation.WithMessage("invalid membership id")
	}
	if !token.ValidateTokenScope(req.TokenScope) {
		return apperror.ErrValidation.WithMessage("invalid token scope")
	}
	if !constants.IsValidaOrgMemStatus(req.NewStatus) {
		return apperror.ErrValidation.WithMessage("invalid status transition")
	}
	if req.TokenScope != string(token.TokenScopeUpdateOrganizationMembership) {
		return apperror.ErrPermissionDenied.WithMessage("you do not have permission to update this membership")
	}
	if req.NewStatus == constants.OrgMemStatusLeft {
		return apperror.ErrValidation.WithMessage("cannot leave an organization through this request")
	}

	return nil
}
