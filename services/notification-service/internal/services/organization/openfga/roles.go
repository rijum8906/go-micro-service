package openfga

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	orgjobsdto "github.com/rijum8906/relay/packages/core/dto/jobs/organization"
	organizationconstants "github.com/rijum8906/relay/services/organization-service/app/constants"
	"go.uber.org/zap"
)

// processUpdateOrgMemRole processes role change events from NATS.
//
// Execution flow:
//  1. Unmarshal the incoming message into UpdatedRoleDTO
//  2. Delete the old role tuple (if exists)
//  3. Create the new role tuple
//  4. Acknowledge successful processing
//
// Error Handling:
//   - Idempotent operations: OpenFGA tuples are idempotent, so retries won't create duplicates
//   - NotFound errors: Treated as success (role already removed)
//   - Other errors: Return to trigger retry mechanism
//   - Unmarshal errors: Return immediately (invalid data can't be retried)
func (s *OrgPermissionService) processUpdateOrgMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.UpdateOrgRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Delete old role
	if err := s.deleteOrgRole(ctx, &data.OrgRoleDTO); err != nil {
		return err
	}

	// Create new role
	if err := s.createOrgRole(ctx, &data.OrgRoleDTO); err != nil {
		return err
	}

	return nil
}

func (s *OrgPermissionService) processAssignOrgMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.OrgRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Create new role
	if err := s.createOrgRole(ctx, &data); err != nil {
		return err
	}

	return nil
}

func (s *OrgPermissionService) processRevokeOrgMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.OrgRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Delete role
	if err := s.deleteOrgRole(ctx, &data); err != nil {
		return err
	}

	return nil
}

func (s *OrgPermissionService) processAssignTeamMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.TeamRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Create new role
	if err := s.assignTeamMembershipRole(ctx, &data); err != nil {
		return err
	}

	return nil
}

func (s *OrgPermissionService) processRevokeTeamMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.TeamRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Create new role
	if err := s.revokeTeamMembershipRole(ctx, &data); err != nil {
		return err
	}

	return nil
}

func (s *OrgPermissionService) processUpdateTeamMemRole(msg *nats.Msg) *apperror.AppError {
	var data orgjobsdto.UpdateTeamRoleDTO

	// Parse message
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.
			WithMessage("failed to unmarshal role update message").
			WithDetail("error", err.Error())
	}

	ctx := context.Background()

	// Delete old role
	if err := s.revokeTeamMembershipRole(ctx, &data.TeamRoleDTO); err != nil {
		return err
	}

	// Create new role
	if err := s.assignTeamMembershipRole(ctx, &data.TeamRoleDTO); err != nil {
		return err
	}

	return nil
}

// deleteOrgRole removes the old role tuple from OpenFGA.
// Returns nil if the role doesn't exist (idempotent).
func (s *OrgPermissionService) deleteOrgRole(ctx context.Context, data *orgjobsdto.OrgRoleDTO) *apperror.AppError {
	var appErr *apperror.AppError

	if organizationconstants.IsStandardOrgRole(data.Role) {
		appErr = s.PermissionManager.DeleteCustomOrgRole(ctx, data.User, data.Organization, data.Role)
	} else {
		appErr = s.PermissionManager.DeleteCustomOrgRole(ctx, data.User, data.Organization, data.Role)
	}

	// NotFound is acceptable - role may have been already removed
	if appErr != nil && appErr.Code == apperror.CodeNotFound {
		s.Logger.Debug(
			"role not found during djeletion, treating as success",
			zap.String("user_id", data.User),
			zap.String("organization_id", data.Organization),
			zap.String("role", data.Role),
		)
		return nil
	}

	return appErr
}

// createOrgRole adds the new role tuple to OpenFGA.
func (s *OrgPermissionService) createOrgRole(ctx context.Context, data *orgjobsdto.OrgRoleDTO) *apperror.AppError {
	if organizationconstants.IsStandardOrgRole(data.Role) {
		return s.PermissionManager.CreateOrgRole(ctx, data.User, data.Organization, data.Role)
	}

	return s.PermissionManager.CreateCustomOrgRole(ctx, data.User, data.Organization, data.Role)
}

func (s *OrgPermissionService) assignTeamMembershipRole(ctx context.Context, data *orgjobsdto.TeamRoleDTO) *apperror.AppError {
	if organizationconstants.IsStandardOrgTeamRole(data.Role) {
		return s.PermissionManager.CreateOrgTeamRole(ctx, data.User, data.Team, data.Role)
	}

	return s.PermissionManager.CreateCustomOrgTeamRole(ctx, data.User, data.Team, data.Role)
}

func (s *OrgPermissionService) revokeTeamMembershipRole(ctx context.Context, data *orgjobsdto.TeamRoleDTO) *apperror.AppError {
	if organizationconstants.IsStandardOrgTeamRole(data.Role) {
		return s.PermissionManager.DeleteOrgTeamRole(ctx, data.User, data.Team, data.Role)
	}

	return s.PermissionManager.DeleteCustomOrgTeamRole(ctx, data.User, data.Team, data.Role)
}
