package openfga

import (
	orgjobsdto "github.com/rijum8906/relay/packages/core/dto/jobs/organization"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func (s *Repository) buildOrgMemRoleDTO(targetMembership *db.OrganizationMembership) orgjobsdto.OrgRoleDTO {
	return orgjobsdto.OrgRoleDTO{
		User:         targetMembership.UserID.String(),
		Role:         targetMembership.Role,
		Organization: targetMembership.OrganizationID.String(),
	}
}

func (s *Repository) buildOrgTeamRoleDTO(targetMembership *db.OrganizationTeamMembership) orgjobsdto.OrgRoleDTO {
	return orgjobsdto.OrgRoleDTO{
		User:         targetMembership.MembershipID.String(),
		Role:         targetMembership.Role,
		Organization: targetMembership.OrganizationID.String(),
	}
}
