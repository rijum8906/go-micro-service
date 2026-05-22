package orgteam

import (
	"context"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

// checkCreateTeamPermission verifies the user has permission to create teams in the organization.
func (s *OrgTeamService) checkCreateTeamPermission(ctx context.Context, userID string, orgID uuid.UUID) *apperror.AppError {
	checkResp, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userID,
		Relation: permissions.PermissionCanCreateTeam,
		Object:   "organization:" + orgID.String(),
	})
	if appErr != nil {
		return apperror.ErrInternal.
			WithMessage("failed to verify team creation permission").
			WithDetail("error", appErr.Error())
	}

	if !*checkResp.Allowed {
		return apperror.ErrPermissionDenied.
			WithMessage("you do not have permission to create teams in this organization").
			WithDetail("user_id", userID).
			WithDetail("organization_id", orgID.String())
	}

	return nil
}

// validateUniqueTeamName checks if a team with the given name already exists in the organization.
func (s *OrgTeamService) validateUniqueTeamName(ctx context.Context, teamName string, orgID uuid.UUID) *apperror.AppError {
	exists, err := s.DBQ.CheckOrganizationTeamNameExists(ctx, db.CheckOrganizationTeamNameExistsParams{
		Name:           teamName,
		OrganizationID: orgID,
	})
	if err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to check team name uniqueness").
			WithDetail("error", err.Error())
	}

	if exists {
		return apperror.ErrValidation.
			WithMessage("team name already exists in this organization").
			WithDetail("team_name", teamName).
			WithDetail("organization_id", orgID.String())
	}

	return nil
}
