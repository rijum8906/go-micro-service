package orgteam

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_teamv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_team/v1"
	"github.com/rijum8906/relay/services/organization-service/app/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

// CreateTeam creates a new team within an organization and assigns the creator as the team owner.
//
// This method creates a team entity and establishes the creator as the team owner with
// full administrative privileges over the team.
//
// Execution Flow:
//   - Validate request parameters (organization ID, team name, description)
//   - Authenticate and extract user identity from context
//   - Verify user has permission to create teams in the organization (via OpenFGA)
//   - Validate team name uniqueness within the organization
//   - Begin database transaction
//   - Create team record
//   - Create team membership record with owner role for the creator
//   - Commit transaction
//   - AFTER successful commit, assign OpenFGA permissions for team ownership
//   - Return success response with team details
//
// Security Constraints:
//   - User must have `PermissionCanCreateTeam` in the organization
//   - Team names must be unique per organization
//   - Creator automatically becomes team owner (RoleOrgTeamOwner)
//   - Team ownership is granted via OpenFGA after successful creation
//
// Idempotency:
//   - Not idempotent - each call creates a new unique team
//   - Team name uniqueness is enforced by database constraint
//   - Duplicate team names return validation error
//
// Database Transaction:
//   - Creates team and membership in a single atomic operation
//   - Ensures both records exist before committing
//   - Rolls back if either creation fails
//
// OpenFGA Strategy:
//   - Permissions assigned AFTER database commit (best-effort)
//   - If OpenFGA fails, team exists but user lacks permissions (requires reconciliation)
//   - Failure is logged and returned as error to trigger retry
//   - This is acceptable because team creation is rare and permissions can be fixed
//
// Why OpenFGA assignment AFTER commit:
//   - Database transaction ensures team and membership are created
//   - Team exists even if permission assignment fails (can be retried)
//   - Prevents orphaned permissions if database creation fails
//   - Clear separation: DB creation is critical, permissions are secondary
//
// Error Responses:
//   - Validation:      Invalid organization ID, team name already exists
//   - PermissionDenied: User lacks team creation permission
//   - NotFound:         Organization or membership not found
//   - Internal:        Database operation failed or OpenFGA assignment failed
//
// Example:
//
//	team, err := service.CreateTeam(ctx, &org_teamv1.CreateOrgTeamRequest{
//	    OrganizationId: orgID,
//	    Name: "Engineering",
//	    Description: "Engineering team responsible for product development",
//	})
//	if err != nil {
//	    return nil, err
//	}
//	fmt.Printf("Created team: %s (ID: %s)", team.Name, team.Id)
func (s *OrgTeamService) CreateTeam(ctx context.Context, req *org_teamv1.CreateOrgTeamRequest) (*org_teamv1.OrgTeamRes, error) {
	// Validate request parameters
	orgID, membershipID, appErr := parseCreateOrgTeamReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("invalid user ID format").WithDetail("error", err.Error())
	}

	// Verify user has permission to create teams in this organization
	if appErr := s.checkCreateTeamPermission(ctx, userInfo.UserID, orgID); appErr != nil {
		return nil, appErr
	}

	// Validate team name uniqueness within the organization
	if appErr := s.validateUniqueTeamName(ctx, req.Name, orgID); appErr != nil {
		return nil, appErr
	}

	// Store created entities for use after transaction
	var team *db.OrganizationTeam
	var teamMembership *db.OrganizationTeamMembership

	// Execute database transaction (team + membership creation)
	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Create the team record
		t, createErr := q.CreateOrganizationTeam(ctx, db.CreateOrganizationTeamParams{
			Description: pgtype.Text{
				String: req.Description,
				Valid:  req.Description != "",
			},
			Name:           req.Name,
			OrganizationID: orgID,
			CreatedByMemID: userID,
		})
		if createErr != nil {
			return apperror.ErrInternal.
				WithMessage("failed to create team").
				WithDetail("error", createErr.Error())
		}
		team = &t

		// Create team membership with owner role for the creator
		tm, createErr := q.CreateOrganizationTeamMembership(ctx, db.CreateOrganizationTeamMembershipParams{
			OrganizationID: orgID,
			Role:           constants.OrgTeamRoleOwner,
			TeamID:         team.ID,
			MembershipID:   membershipID,
		})
		if createErr != nil {
			return apperror.ErrInternal.
				WithMessage("failed to create team membership").
				WithDetail("error", createErr.Error())
		}
		teamMembership = &tm

		s.Logger.Info("Team and membership created successfully",
			zap.String("team_id", team.ID.String()),
			zap.String("team_name", team.Name),
			zap.String("organization_id", orgID.String()),
			zap.String("created_by", userInfo.UserID))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// AFTER successful transaction, assign OpenFGA team owner permissions (best-effort)
	if appErr := s.OpenFGARepo.PublishAssignTeamMemRoleJob(ctx, teamMembership); appErr != nil {
		// CRITICAL: Team created but OpenFGA sync failed
		s.Logger.Error("CRITICAL: Failed to assign OpenFGA team permissions after team creation",
			zap.String("team_id", team.ID.String()),
			zap.String("team_name", team.Name),
			zap.String("organization_id", orgID.String()),
			zap.String("user_id", userInfo.UserID),
			zap.String("membership_id", membershipID.String()),
			zap.Error(appErr))

		// Return error to notify client - team exists but needs permission fix
		return nil, apperror.ErrInternal.
			WithMessage("team created but permission assignment failed").
			WithDetail("reason", "Please contact support or retry the operation").
			WithDetail("team_id", team.ID.String())
	}

	s.Logger.Info("OpenFGA team permissions assigned successfully",
		zap.String("team_id", team.ID.String()),
		zap.String("user_id", userInfo.UserID))

	// Return success response
	return &org_teamv1.OrgTeamRes{
		Id:             team.ID.String(),
		Name:           team.Name,
		Description:    team.Description.String,
		OrganizationId: team.OrganizationID.String(),
	}, nil
}

func (s *OrgTeamService) UpdateTeamName(ctx context.Context, req *org_teamv1.UpdateOrgTeamNameRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *OrgTeamService) DeleteTeam(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
