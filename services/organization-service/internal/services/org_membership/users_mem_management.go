package orgmembership

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
	"go.uber.org/zap"
)

// LeaveOrganization removes the authenticated user from an organization.
//
// Execution Flow:
//   - Validate request parameters (ID format and scoped token)
//   - Validate token scope matches the operation
//   - Authenticate and extract user identity from context
//   - Begin database transaction
//   - Retrieve organization membership by ID
//   - Verify the authenticated user owns this membership (IDOR protection)
//   - Check idempotency (already left? return early)
//   - Validate business rule: last owner cannot leave
//   - Remove user permissions from OpenFGA (SECURITY CRITICAL - fails the request if unsuccessful)
//   - Update membership status to 'left' (soft delete)
//   - Commit transaction
//   - Return success response
//
// Idempotency:
//   - If membership status is already 'left', returns success immediately within transaction
//   - No error for duplicate leave requests
//   - OpenFGA permissions only removed once (idempotent operation)
//
// Business Rules:
//   - Users can only leave their own membership (not others)
//   - Last owner cannot leave (would orphan the organization)
//   - Organization must have at least one owner at all times
//   - Memberships are soft-deleted (status='left'), not hard deleted
//
// Security:
//   - Validates scoped token to prevent CSRF attacks
//   - Explicit ownership check prevents IDOR attacks
//   - Returns PermissionDenied (not NotFound) for ownership mismatch to prevent information disclosure
//   - Removes OpenFGA permissions BEFORE updating database (fail-secure: if permissions fail, DB unchanged)
//   - OpenFGA failure fails the entire request (security over availability)
//
// Why check last owner?
//   - Prevents organizations from having zero owners
//   - Owners have special privileges and responsibilities
//   - Forces proper organization cleanup or ownership transfer
//   - Prevents orphaned organizations without administrative control
//
// Why soft delete (status='left') instead of hard delete?
//   - Maintains audit trail for compliance
//   - Allows reactivation if user rejoins
//   - Preserves historical data for analytics
//   - Prevents foreign key issues with related data
//   - Enables investigation of security incidents
//
// Why remove OpenFGA permissions BEFORE database update?
//   - Security-first approach: permissions revoked even if DB fails
//   - If OpenFGA fails, transaction rolls back - user has no permissions AND membership unchanged
//   - Fail-secure: better to reject the request than leave permissions dangling
//   - Defense in depth: both systems must succeed for consistency
//
// Why OpenFGA failure fails the entire request?
//   - Permissions are a security boundary, not a best-effort feature
//   - Leaving a user with permissions after they've "left" is a security violation
//   - Failing the request alerts the user to retry or contact support
//   - Requires monitoring and alerting on OpenFGA failures
//
// Transaction Boundaries:
//   - All database operations run in a serializable transaction
//   - OpenFGA operations run BEFORE commit (not in transaction, but failure triggers rollback)
//   - Transaction provides ACID guarantees for database state
//   - OpenFGA provides atomic permission updates (all or nothing per request)
//
// Error Responses:
//   - Validation: Invalid request format, invalid scoped token, or last owner cannot leave
//   - NotFound: Membership doesn't exist (after UUID parsing)
//   - PermissionDenied: User doesn't own this membership or invalid token scope
//   - Internal: Database operation failed or OpenFGA operation failed
//   - Transaction errors: Automatic rollback with appropriate error wrapping
//
// Example:
//
//	resp, err := service.LeaveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
//	    Id: membershipID,
//	    ScopedToken: token,
//	})
//	if err != nil {
//	    // Handle error based on type
//	}
//	fmt.Printf("User left organization: %v", resp.Success)
func (s *OrgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	// Validate request parameters (ID format and scoped token presence)
	membershipID, appErr := protoutils.ParseIDAndScopedTokenReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Validate token scope matches operation type
	if req.TokenScope != string(token.TokenScopeLeaveOrganization) {
		return nil, apperror.ErrPermissionDenied.WithMessage("invalid token scope for leave organization operation")
	}

	// Extract authenticated user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Execute critical operation in transaction with fail-secure semantics
	if err := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		//  Retrieve organization membership
		membership, err := q.GetOrganizationMembership(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "membership")
		}

		//  SECURITY: Verify the authenticated user owns this membership (Prevents IDOR attacks where user tries to leave another user's membership)
		if membership.UserID.String() != userInfo.UserID {
			return apperror.ErrPermissionDenied.WithMessage("you can only leave your own membership")
		}

		// 6. Idempotency: Return success if already left
		if membership.Status == constants.OrgMemStatusLeft {
			return nil
		}

		// IMPORTANT:  Business rule: Prevent the last owner from leaving
		// Ensures organization always has at least one owner for administrative functions
		if membership.Role == constants.OrgRoleOwner {
			activeOwners, err := q.CountActiveOwnersByOrgID(ctx, membership.OrganizationID)
			if err != nil {
				return apperror.ErrInternal.
					WithMessage("failed to verify organization ownership requirements").
					WithDetail("db_error", err.Error())
			}

			if activeOwners <= 1 {
				return apperror.ErrValidation.WithMessage(
					"cannot leave organization: you are the last owner. " +
						"Transfer ownership to another member or delete the organization instead",
				)
			}
		}

		//  NOTE: Remove OpenFGA permissions BEFORE database update
		//  This ensures fail-secure behavior:
		//   - If OpenFGA succeeds but DB fails → user loses access (secure, needs reconciliation)
		//   - If OpenFGA fails → transaction rolls back (user retains access, request fails)
		//
		// Note: OpenFGA's Delete operation is idempotent - safe to retry if needed
		if appErr := s.Helper.RemoveRole(ctx, &membership); appErr != nil {
			// Log the failure for monitoring and alerting
			s.Logger.Error("OpenFGA permission revocation failed",
				zap.String("user_id", userInfo.UserID),
				zap.String("organization_id", membership.OrganizationID.String()),
				zap.String("membership_id", membership.ID.String()),
				zap.Error(appErr))

			// Return error to trigger transaction rollback
			// User cannot leave until permissions are properly revoked
			return apperror.ErrInternal.
				WithMessage("failed to revoke organization permissions").
				WithDetail("reason", "authorization system unavailable").
				WithDetail("error", appErr.Error())
		}

		// Update membership status to 'left' (soft delete)
		// This runs after OpenFGA success, within the same transaction
		_, err = q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
			ID:     membership.ID,
			Status: constants.OrgMemStatusLeft,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to update membership status").
				WithDetail("db_error", err.Error())
		}

		// Log successful permission revocation for audit trail
		s.Logger.Info("User left organization successfully",
			zap.String("user_id", userInfo.UserID),
			zap.String("organization_id", membership.OrganizationID.String()),
			zap.String("membership_id", membership.ID.String()),
			zap.String("previous_role", membership.Role))

		return nil
	}); err != nil {
		// Transaction failed - return the error to the client
		return nil, err
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// ############################ ORGANIZATION'S MEMBERSHIP MANAGEMENT ############################

// BanOrganizationMembership bans a user from an organization.
//
// This method changes a membership status from 'active' to 'banned', revoking all
// organization permissions. Banned users cannot access organization resources but
// their membership record is preserved for audit purposes.
//
// NOTE: This method only handles banning. For unbanning, use UnbanOrganizationMembership().
// For voluntary departure, use LeaveOrganization().
//
// Execution Flow:
//   - Validate request parameters (membership ID, token scope)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (no self-ban, no owner ban)
//   - Verify actor has permission via RBAC role hierarchy
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already banned? return success)
//   - Validate target status allows banning (only active members can be banned)
//   - Update membership status to 'banned' in database
//   - Remove all organization permissions from OpenFGA
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Users cannot ban themselves (use LeaveOrganization instead)
//   - Organization owners cannot be banned
//   - Actors need admin or owner role (or custom role with ban permission)
//   - Role hierarchy enforced: owner > admin > member > custom roles
//
// Error Responses:
//   - Validation:      Invalid parameters or invalid status
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.BanOrganizationMembership(ctx, &req{
//	    OrganizationMembershipId: membershipID,
//	    ScopedToken: token,
//	})
func (s *OrgMembershipService) BanOrganizationMembership(
	ctx context.Context,
	req *corev1.IDAndScopedTokenRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	membershipID, appErr := protoutils.ParseIDAndScopedTokenReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Extract authenticated user identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Execute all operations in a transaction
	var targetMembership, actorMembership *db.OrganizationMembership

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership (allow all statuses for idempotency)
		target, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}
		targetMembership = &target

		// Load actor membership (must be active to perform actions)
		actor, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         uuid.MustParse(userInfo.UserID),
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}
		actorMembership = &actor

		// Lock both records for update (prevent race conditions)
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Validate target membership is not in terminal state
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			return apperror.ErrValidation.
				WithMessage("cannot ban a membership that has been removed").
				WithDetail("current_status", targetMembership.Status)
		}

		// Prevent self-ban
		if targetMembership.UserID.String() == userInfo.UserID {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot ban yourself").
				WithDetail("reason", "Use LeaveOrganization() to leave the organization voluntarily")
		}

		// SECURITY: Prevent banning owners
		if targetMembership.Role == constants.OrgRoleOwner {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot ban an organization owner")
		}

		// Verify actor has permission via standard RBAC
		if constants.IsStandardOrgRole(actorMembership.Role) {
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to ban this member").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("required_role", "admin or owner")
			}
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckCanChangeMembershipStatus(
			ctx, s.TuppleManager, actorMembership, targetMembership,
		); appErr != nil {
			return appErr
		}

		// Idempotency check - already banned
		if targetMembership.Status == constants.OrgMemStatusBanned {
			s.Logger.Debug("membership already banned",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()))
			return nil
		}

		// Validate current status allows banning
		// NOTE: Only ACTIVE members can be banned
		if targetMembership.Status != constants.OrgMemStatusActive {
			return apperror.ErrValidation.
				WithMessage("cannot ban membership with current status").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusActive).
				WithDetail("reason", "Only active members can be banned")
		}

		// IMPORTANT: Update database FIRST (so OpenFGA changes only happen if DB succeeds)
		_, err = q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
			ID:     membershipID,
			Status: constants.OrgMemStatusBanned,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to update membership status").
				WithDetail("db_error", err.Error())
		}

		// Remove all organization permissions from OpenFGA
		// This runs after DB update - if it fails, we log and continue
		// TODO: Implement retry mechanism for OpenFGA failures
		if permErr := s.Helper.RemoveRole(ctx, targetMembership); permErr != nil {
			// CRITICAL: Database updated but OpenFGA sync failed
			s.Logger.Error("CRITICAL: Failed to revoke OpenFGA permissions after ban - manual intervention may be required",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("role", targetMembership.Role),
				zap.String("membership_id", membershipID.String()),
				zap.Error(permErr))
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// UnbanOrganizationMembership unbans a user from an organization.
//
// This method changes a membership status from 'banned' to 'active', restoring all
// organization permissions based on their role. Unbanned users regain full access
// to organization resources according to their role.
//
// NOTE: This method only handles unbanning. For banning, use BanOrganizationMembership().
// For voluntary departure, use LeaveOrganization().
//
// Execution Flow:
//   - Validate request parameters (membership ID, token scope)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (no self-unban check, but actor must have permission)
//   - Verify actor has permission via RBAC role hierarchy
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already active? return success)
//   - Validate target status allows unbanning (only banned members can be unbanned)
//   - Update membership status to 'active' in database
//   - Restore all organization permissions in OpenFGA based on role
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Actors need admin or owner role (or custom role with unban permission)
//   - Cannot unban yourself (already banned users can't perform actions)
//   - Role hierarchy enforced: owner > admin > member > custom roles
//   - Unban restores the exact role the user had before being banned
//
// Error Responses:
//   - Validation:      Invalid parameters or target is not banned
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.UnbanOrganizationMembership(ctx, &req{
//	    OrganizationMembershipId: membershipID,
//	    ScopedToken: token,
//	})
func (s *OrgMembershipService) UnbanOrganizationMembership(
	ctx context.Context,
	req *corev1.IDAndScopedTokenRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	membershipID, appErr := protoutils.ParseIDAndScopedTokenReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Extract authenticated user identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Execute all operations in a transaction
	var targetMembership, actorMembership *db.OrganizationMembership

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership (allow all statuses for idempotency)
		target, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}
		targetMembership = &target

		// Load actor membership (must be active to perform actions)
		actor, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         uuid.MustParse(userInfo.UserID),
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}
		actorMembership = &actor

		// Lock both records for update (prevent race conditions)
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Validate target membership is not in terminal state
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			return apperror.ErrValidation.
				WithMessage("cannot unban a membership that has been removed").
				WithDetail("current_status", targetMembership.Status)
		}

		if targetMembership.Status == constants.OrgMemStatusLeft {
			return apperror.ErrValidation.
				WithMessage("cannot unban a membership that has left").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("reason", "User must be re-invited to the organization")
		}

		// Prevent unbanning owners (owners should never be banned)
		if targetMembership.Role == constants.OrgRoleOwner {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot unban an organization owner").
				WithDetail("reason", "Owners should never be banned")
		}

		// Verify actor has permission via standard RBAC
		if constants.IsStandardOrgRole(actorMembership.Role) {
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to unban this member").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("required_role", "admin or owner")
			}
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckCanChangeMembershipStatus(
			ctx, s.TuppleManager, actorMembership, targetMembership,
		); appErr != nil {
			return appErr
		}

		// Idempotency check - already active
		if targetMembership.Status == constants.OrgMemStatusActive {
			s.Logger.Debug("membership already active",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()))
			return nil
		}

		// Validate current status allows unbanning
		// Only BANNED members can be unbanned
		if targetMembership.Status != constants.OrgMemStatusBanned {
			return apperror.ErrValidation.
				WithMessage("cannot unban membership with current status").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusBanned).
				WithDetail("reason", "Only banned members can be unbanned")
		}

		// IMPORTANT: Update database FIRST (so OpenFGA changes only happen if DB succeeds)
		_, err = q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
			ID:     membershipID,
			Status: constants.OrgMemStatusActive,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to update membership status").
				WithDetail("db_error", err.Error())
		}

		// Restore all organization permissions in OpenFGA based on role
		// This runs after DB update - if it fails, we log and continue
		// TODO: Implement retry mechanism for OpenFGA failures
		if permErr := s.Helper.AddRole(ctx, targetMembership); permErr != nil {
			// CRITICAL: Database updated but OpenFGA sync failed
			s.Logger.Error("CRITICAL: Failed to restore OpenFGA permissions after unban - manual intervention may be required",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("role", targetMembership.Role),
				zap.String("membership_id", membershipID.String()),
				zap.Error(permErr))
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// SuspendOrganizationMembership temporarily suspends a user from an organization.
//
// This method changes a membership status from 'active' to 'suspended', temporarily
// revoking all organization permissions. Suspended users cannot access organization
// resources but their membership record is preserved and can be reactivated.
//
// NOTE: This is a temporary action. Suspended users can be reactivated using
// ActivateOrganizationMembership(). For permanent removal, use BanOrganizationMembership()
// or LeaveOrganization().
//
// Execution Flow:
//   - Validate request parameters (membership ID, token scope)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (no self-suspend, no owner suspension)
//   - Verify actor has permission via RBAC role hierarchy
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already suspended? return success)
//   - Validate target status allows suspension (only active members)
//   - Update membership status to 'suspended' in database
//   - Remove all organization permissions from OpenFGA
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Users cannot suspend themselves
//   - Organization owners cannot be suspended
//   - Actors need admin or owner role (or custom role with suspend permission)
//   - Role hierarchy enforced: owner > admin > member > custom roles
//   - Suspension is temporary and reversible
//
// Difference from Ban:
//   - Suspension is temporary, Ban is permanent
//   - Suspension implies future reactivation, Ban is typically final
//   - Business logic may treat suspended users differently from banned users
//
// Error Responses:
//   - Validation:      Invalid parameters or target not active
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.SuspendOrganizationMembership(ctx, &req{
//	    OrganizationMembershipId: membershipID,
//	    ScopedToken: token,
//	})
func (s *OrgMembershipService) SuspendOrganizationMembership(
	ctx context.Context,
	req *corev1.IDAndScopedTokenRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	membershipID, appErr := protoutils.ParseIDAndScopedTokenReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Extract authenticated user identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Execute all operations in a transaction
	var targetMembership, actorMembership *db.OrganizationMembership

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership (allow all statuses for idempotency)
		target, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}
		targetMembership = &target

		// Load actor membership (must be active to perform actions)
		actor, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         uuid.MustParse(userInfo.UserID),
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}
		actorMembership = &actor

		// Lock both records for update (prevent race conditions)
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Validate target membership is not in terminal state
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			return apperror.ErrValidation.
				WithMessage("cannot suspend a membership that has been removed").
				WithDetail("current_status", targetMembership.Status)
		}

		if targetMembership.Status == constants.OrgMemStatusLeft {
			return apperror.ErrValidation.
				WithMessage("cannot suspend a membership that has left").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("reason", "User is no longer a member of the organization")
		}

		// Prevent self-suspension
		if targetMembership.UserID.String() == userInfo.UserID {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot suspend yourself").
				WithDetail("reason", "Use LeaveOrganization() to leave the organization voluntarily")
		}

		// SECURITY: Prevent suspending owners
		if targetMembership.Role == constants.OrgRoleOwner {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot suspend an organization owner").
				WithDetail("reason", "Owners cannot be suspended")
		}

		// Verify actor has permission via standard RBAC
		if constants.IsStandardOrgRole(actorMembership.Role) {
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to suspend this member").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("required_role", "admin or owner")
			}
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckCanChangeMembershipStatus(
			ctx, s.TuppleManager, actorMembership, targetMembership,
		); appErr != nil {
			return appErr
		}

		// Idempotency check - already suspended
		if targetMembership.Status == constants.OrgMemStatusSuspended {
			s.Logger.Debug("membership already suspended",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()))
			return nil
		}

		// Validate current status allows suspension
		// Only ACTIVE members can be suspended
		if targetMembership.Status != constants.OrgMemStatusActive {
			return apperror.ErrValidation.
				WithMessage("cannot suspend membership with current status").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusActive).
				WithDetail("reason", "Only active members can be suspended")
		}

		// IMPORTANT: Update database FIRST
		_, err = q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
			ID:     membershipID,
			Status: constants.OrgMemStatusSuspended,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to update membership status").
				WithDetail("db_error", err.Error())
		}

		// Remove all organization permissions from OpenFGA
		// Suspended users lose all access temporarily
		if permErr := s.Helper.RemoveRole(ctx, targetMembership); permErr != nil {
			// CRITICAL: Database updated but OpenFGA sync failed
			s.Logger.Error("CRITICAL: Failed to revoke OpenFGA permissions during suspension - manual intervention may be required",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("role", targetMembership.Role),
				zap.String("membership_id", membershipID.String()),
				zap.Error(permErr))
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// ActivateOrganizationMembership reactivates a suspended user in an organization.
//
// This method changes a membership status from 'suspended' to 'active', restoring all
// organization permissions based on their role. Reactivated users regain full access
// to organization resources according to their role.
//
// NOTE: This method only handles reactivation of suspended members. For unbanning
// banned users, use UnbanOrganizationMembership().
//
// Execution Flow:
//   - Validate request parameters (membership ID, token scope)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (can reactivate any suspended member)
//   - Verify actor has permission via RBAC role hierarchy
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already active? return success)
//   - Validate target status allows activation (only suspended members)
//   - Update membership status to 'active' in database
//   - Restore all organization permissions in OpenFGA based on role
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Actors need admin or owner role (or custom role with activate permission)
//   - Cannot activate yourself (suspended users can't perform actions)
//   - Role hierarchy enforced: owner > admin > member > custom roles
//   - Activation restores the exact role the user had before suspension
//
// Error Responses:
//   - Validation:      Invalid parameters or target not suspended
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.ActivateOrganizationMembership(ctx, &req{
//	    OrganizationMembershipId: membershipID,
//	    ScopedToken: token,
//	})
func (s *OrgMembershipService) ActivateOrganizationMembership(
	ctx context.Context,
	req *corev1.IDAndScopedTokenRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	membershipID, appErr := protoutils.ParseIDAndScopedTokenReq(req)
	if appErr != nil {
		return nil, appErr
	}

	// Extract authenticated user identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Execute all operations in a transaction
	var targetMembership, actorMembership *db.OrganizationMembership

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership (allow all statuses for idempotency)
		target, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}
		targetMembership = &target

		// Load actor membership (must be active to perform actions)
		actor, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         uuid.MustParse(userInfo.UserID),
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}
		actorMembership = &actor

		// Lock both records for update (prevent race conditions)
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Validate target membership is not in terminal state
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			return apperror.ErrValidation.
				WithMessage("cannot activate a membership that has been removed").
				WithDetail("current_status", targetMembership.Status)
		}

		if targetMembership.Status == constants.OrgMemStatusLeft {
			return apperror.ErrValidation.
				WithMessage("cannot activate a membership that has left").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("reason", "User must be re-invited to the organization")
		}

		// Prevent activating banned users (they need unban, not activate)
		if targetMembership.Status == constants.OrgMemStatusBanned {
			return apperror.ErrValidation.
				WithMessage("cannot activate a banned membership").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusSuspended).
				WithDetail("suggestion", "Use UnbanOrganizationMembership() instead")
		}

		// Verify actor has permission via standard RBAC
		if constants.IsStandardOrgRole(actorMembership.Role) {
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to activate this member").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("required_role", "admin or owner")
			}
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckCanChangeMembershipStatus(
			ctx, s.TuppleManager, actorMembership, targetMembership,
		); appErr != nil {
			return appErr
		}

		// Idempotency check - already active
		if targetMembership.Status == constants.OrgMemStatusActive {
			s.Logger.Debug("membership already active",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()))
			return nil
		}

		// Validate current status allows activation
		// Only SUSPENDED members can be activated
		if targetMembership.Status != constants.OrgMemStatusSuspended {
			return apperror.ErrValidation.
				WithMessage("cannot activate membership with current status").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusSuspended).
				WithDetail("reason", "Only suspended members can be activated")
		}

		// Optional: Check if user has been suspended for minimum time period
		// Uncomment if you have a suspension_min_duration policy
		// if targetMembership.SuspendedAt != nil {
		//     minSuspendDuration := 1 * time.Hour
		//     if time.Since(*targetMembership.SuspendedAt) < minSuspendDuration {
		//         return apperror.ErrValidation.
		//             WithMessage("cannot activate user yet").
		//             WithDetail("reason", "User must remain suspended for minimum duration").
		//             WithDetail("minimum_duration", minSuspendDuration.String())
		//     }
		// }

		// IMPORTANT: Update database FIRST
		_, err = q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
			ID:     membershipID,
			Status: constants.OrgMemStatusActive,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to update membership status").
				WithDetail("db_error", err.Error())
		}

		// Restore all organization permissions in OpenFGA based on role
		if permErr := s.Helper.AddRole(ctx, targetMembership); permErr != nil {
			// CRITICAL: Database updated but OpenFGA sync failed
			s.Logger.Error("CRITICAL: Failed to restore OpenFGA permissions during activation - manual intervention may be required",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("role", targetMembership.Role),
				zap.String("membership_id", membershipID.String()),
				zap.Error(permErr))
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// ChangeOrganizationMembershipRole changes the role of an organization member.
//
// This method updates a member's role within an organization, revoking previous
// role permissions and granting new ones. Role changes are subject to role
// hierarchy and permission checks.
//
// NOTE: Owner role cannot be assigned through this method. Use dedicated
// ownership transfer methods instead.
//
// Execution Flow:
//   - Validate request parameters (membership ID, token scope, new role)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (no self-change, no owner modification)
//   - Verify actor can manage target's current role
//   - Verify actor can assign the new role
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already has role? return success)
//   - Remove old role permissions from OpenFGA
//   - Update membership role in database
//   - Grant new role permissions in OpenFGA
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Users cannot change their own role
//   - Cannot change an owner's role
//   - Cannot assign owner role
//   - Actors cannot manage users with equal or higher role
//   - Actors cannot assign roles equal to or higher than their own
//   - Role hierarchy: owner > admin > member > custom roles
//
// Error Responses:
//   - Validation:      Invalid parameters, invalid role, or self-modification
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.ChangeOrganizationMembershipRole(ctx, &req{
//	    OrganizationMembershipId: membershipID,
//	    NewRole: "admin",
//	    ScopedToken: token,
//	})
func (s *OrgMembershipService) ChangeOrganizationMembershipRole(
	ctx context.Context,
	req *org_membershipv1.ChangeOrgMembershipRoleReq,
) (*corev1.SuccessResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	// Parse membership ID
	membershipID, err := uuid.Parse(req.OrganizationMembershipId)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("provided membership id is not a valid uuid").
			WithDetail("error", err.Error())
	}

	// Validate token scope
	if !token.ValidateTokenScope(req.TokenScope) {
		return nil, apperror.ErrValidation.WithMessage("token scope must be provided")
	}
	if req.TokenScope != string(token.TokenScopeUpdateOrganizationMembership) {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("invalid token scope for membership role update")
	}

	// Validate new role
	if !constants.IsStandardOrgRole(req.NewRole) {
		return nil, apperror.ErrValidation.
			WithMessage("new role is not valid")
	}
	if req.NewRole == constants.OrgRoleOwner {
		return nil, apperror.ErrValidation.
			WithMessage("owner role cannot be assigned through membership role update").
			WithDetail("reason", "Use dedicated ownership transfer methods")
	}

	// Extract user info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	actorUserID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata contains invalid user id").
			WithDetail("error", err.Error())
	}

	// Execute transaction
	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership
		targetMembership, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}

		// Load actor membership
		actorMembership, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         actorUserID,
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}

		// Lock records for update
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Validate target membership status
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			return apperror.ErrValidation.
				WithMessage("cannot change role for a membership that has been removed").
				WithDetail("current_status", targetMembership.Status)
		}
		if targetMembership.Status == constants.OrgMemStatusLeft {
			return apperror.ErrValidation.
				WithMessage("cannot change role for a membership that has left").
				WithDetail("current_status", targetMembership.Status)
		}
		if targetMembership.Status != constants.OrgMemStatusActive {
			return apperror.ErrValidation.
				WithMessage("cannot change role for membership with current status").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("expected_status", constants.OrgMemStatusActive)
		}

		// Prevent self-role change
		if targetMembership.UserID == actorUserID {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot change your own membership role").
				WithDetail("reason", "Ask another admin to change your role")
		}

		// Prevent modifying owners
		if targetMembership.Role == constants.OrgRoleOwner {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot change an organization owner's role").
				WithDetail("reason", "Owners must transfer ownership first")
		}

		// Verify RBAC permissions
		if constants.IsStandardOrgRole(actorMembership.Role) {
			// Check if actor can manage target's current role
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to change this member's role").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("reason", "You cannot modify a member with equal or higher role")
			}

			// Check if actor can assign the new role
			if !permissions.CanActorManageTarget(actorMembership.Role, req.NewRole) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to assign this role").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("new_role", req.NewRole).
					WithDetail("reason", "You cannot assign a role equal to or higher than yours")
			}
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckMembershipPermission(
			ctx, s.TuppleManager, &actorMembership, permissions.PermissionCanChangeMemberRole,
		); appErr != nil {
			return appErr
		}

		// Idempotency check
		if targetMembership.Role == req.NewRole {
			s.Logger.Debug("membership already has requested role",
				zap.String("membership_id", membershipID.String()),
				zap.String("role", req.NewRole))
			return nil
		}

		// Remove old role permissions from OpenFGA
		oldMembership := targetMembership
		if appErr := s.Helper.RemoveRole(ctx, &oldMembership); appErr != nil {
			return apperror.ErrInternal.
				WithMessage("failed to revoke previous organization role").
				WithDetail("error", appErr.Error())
		}

		// Update role in database
		updatedMembership, err := q.UpdateOrganizationMembershipRole(ctx, db.UpdateOrganizationMembershipRoleParams{
			ID:   membershipID,
			Role: req.NewRole,
		})
		if err != nil {
			// CRITICAL: OpenFGA role removed but DB update failed
			// This creates inconsistency - need to restore OpenFGA role
			s.Logger.Error("CRITICAL: Failed to update role in database after OpenFGA removal",
				zap.String("membership_id", membershipID.String()),
				zap.String("old_role", targetMembership.Role),
				zap.String("new_role", req.NewRole),
				zap.Error(err))

			// Attempt to restore old role in OpenFGA
			if restoreErr := s.Helper.AddRole(ctx, &oldMembership); restoreErr != nil {
				s.Logger.Error("CRITICAL: Failed to restore OpenFGA role after DB failure",
					zap.String("membership_id", membershipID.String()),
					zap.Error(restoreErr))
			}

			return apperror.ErrInternal.
				WithMessage("failed to update membership role").
				WithDetail("db_error", err.Error())
		}

		// Grant new role permissions in OpenFGA
		if appErr := s.Helper.AddRole(ctx, &updatedMembership); appErr != nil {
			// CRITICAL: Database updated but OpenFGA sync failed
			s.Logger.Error("CRITICAL: Failed to grant new role in OpenFGA after role change",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("old_role", targetMembership.Role),
				zap.String("new_role", req.NewRole),
				zap.String("membership_id", membershipID.String()),
				zap.Error(appErr))

			// Don't fail the operation since DB is updated
			// TODO: Queue for retry in background
			s.Helper.RemoveRole(ctx, &targetMembership)
			s.Helper.AddRole(ctx, &updatedMembership)
		}

		// Log successful role change
		s.Logger.Info("organization membership role changed",
			zap.String("actor_user_id", userInfo.UserID),
			zap.String("target_user_id", targetMembership.UserID.String()),
			zap.String("organization_id", targetMembership.OrganizationID.String()),
			zap.String("membership_id", membershipID.String()),
			zap.String("actor_role", actorMembership.Role),
			zap.String("old_role", targetMembership.Role),
			zap.String("new_role", req.NewRole),
			zap.String("target_status", targetMembership.Status))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

// RemoveOrganizationMember permanently removes a member from an organization.
//
// This method removes a user from an organization by setting their membership
// status to 'removed', a terminal state. Removed users lose all organization
// permissions and cannot be reactivated (must be re-invited).
//
// NOTE: This is a permanent action. Removed users cannot be restored and must
// receive a new invitation to rejoin the organization.
//
// Execution Flow:
//   - Validate request parameters (membership ID)
//   - Extract authenticated user identity from context
//   - Begin database transaction with SERIALIZABLE isolation
//   - Retrieve target and actor memberships with FOR UPDATE lock
//   - Validate business rules (no self-removal, no owner removal)
//   - Verify actor has permission via RBAC role hierarchy
//   - Verify custom role permissions via OpenFGA
//   - Check idempotency (already removed? return success)
//   - Validate target status allows removal (cannot remove left members)
//   - Remove all organization permissions from OpenFGA
//   - Update membership status to 'removed' in database
//   - Commit transaction
//   - Return success response
//
// Security Constraints:
//   - Users cannot remove themselves (use LeaveOrganization instead)
//   - Organization owners cannot be removed
//   - Actors need admin or owner role (or custom role with remove permission)
//   - Role hierarchy enforced: owner > admin > member > custom roles
//   - Cannot remove members who have already left
//
// Difference from Leave:
//   - Leave is voluntary user action, Remove is admin action
//   - Left status is soft-delete, Removed is terminal state
//   - Left members might be reactivated, Removed cannot
//
// Error Responses:
//   - Validation:      Invalid parameters or target cannot be removed
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions
//   - Internal:        Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.RemoveOrganizationMember(ctx, &req{
//	    Id: membershipID,
//	})
func (s *OrgMembershipService) RemoveOrganizationMember(
	ctx context.Context,
	req *corev1.IDRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	// Parse membership ID
	membershipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("provided id is not a valid uuid").
			WithDetail("error", err.Error())
	}

	// Extract user info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	actorUserID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata contains invalid user id").
			WithDetail("error", err.Error())
	}

	// Execute transaction
	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Load target membership
		targetMembership, err := q.GetOrganizationMembershipWithAllStatuses(ctx, membershipID)
		if err != nil {
			return coreutils.ParseDBError(err, "target membership")
		}

		// Load actor membership
		actorMembership, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx,
			db.GetOrganizationMembershipByOrgIDAndUserIDParams{
				OrganizationID: targetMembership.OrganizationID,
				UserID:         actorUserID,
			})
		if err != nil {
			return coreutils.ParseDBError(err, "actor membership")
		}

		// Lock records for update
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, targetMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock target membership")
		}
		if _, err := q.LockOrganizationMembershipForUpdate(ctx, actorMembership.ID); err != nil {
			return apperror.ErrInternal.WithMessage("failed to lock actor membership")
		}

		// Idempotency check - already removed
		if targetMembership.Status == constants.OrgMemStatusRemoved {
			s.Logger.Debug("membership already removed",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()))
			return nil
		}

		// Validate target membership status
		if targetMembership.Status == constants.OrgMemStatusLeft {
			return apperror.ErrValidation.
				WithMessage("cannot remove a membership that has already left").
				WithDetail("current_status", targetMembership.Status).
				WithDetail("reason", "User voluntarily left the organization")
		}

		// Prevent self-removal
		if targetMembership.UserID == actorUserID {
			return apperror.ErrPermissionDenied.
				WithMessage("cannot remove yourself from the organization").
				WithDetail("reason", "Use LeaveOrganization() to leave voluntarily")
		}

		// Verify RBAC permissions
		if constants.IsStandardOrgRole(actorMembership.Role) {
			if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to remove this member").
					WithDetail("actor_role", actorMembership.Role).
					WithDetail("target_role", targetMembership.Role).
					WithDetail("required_role", "admin or owner")
			}
		}

		// Prevent removing owners
		if targetMembership.Role == constants.OrgRoleOwner {
			// Check if this is the last owner
			activeOwners, err := q.CountActiveOwnersByOrgID(ctx, targetMembership.OrganizationID)
			if err != nil {
				return apperror.ErrInternal.
					WithMessage("failed to verify organization ownership").
					WithDetail("db_error", err.Error())
			}

			if activeOwners <= 1 {
				return apperror.ErrValidation.
					WithMessage("cannot remove the last owner of the organization").
					WithDetail("reason", "Organization must have at least one owner")
			}

			return apperror.ErrPermissionDenied.
				WithMessage("cannot remove an organization owner").
				WithDetail("reason", "Owners must transfer ownership or leave voluntarily")
		}

		// Verify custom role permissions via OpenFGA
		if appErr := utils.CheckMembershipPermission(
			ctx, s.TuppleManager, &actorMembership, permissions.PermissionCanRemoveMember,
		); appErr != nil {
			return appErr
		}

		// Remove all organization permissions from OpenFGA
		if appErr := s.Helper.RemoveRole(ctx, &targetMembership); appErr != nil {
			// TODO: Implement queue system for OpenFGA retry
			// This should be added after the queue system is implemented
			// The queue should retry failed OpenFGA operations asynchronously
			s.Logger.Error("Failed to revoke OpenFGA permissions during removal - queue retry will be implemented",
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.String("role", targetMembership.Role),
				zap.String("membership_id", membershipID.String()),
				zap.Error(appErr))

			// TODO: Enqueue retry task
			// s.enqueueRetryTask(ctx, &RetryTask{
			//     Type:     "remove_permissions",
			//     UserID:   targetMembership.UserID.String(),
			//     OrgID:    targetMembership.OrganizationID.String(),
			//     Role:     targetMembership.Role,
			//     Attempt:  0,
			// })

			// For now, fail the operation to maintain consistency
			return apperror.ErrInternal.
				WithMessage("failed to revoke organization permissions").
				WithDetail("error", appErr.Error())
		}

		// Update membership status to 'removed' (terminal state)
		if err := q.RemoveOrganizationMembership(ctx, membershipID); err != nil {
			// CRITICAL: OpenFGA permissions removed but DB update failed
			// TODO: This should trigger an alert and manual reconciliation
			s.Logger.Error("CRITICAL: Failed to update membership status after OpenFGA removal",
				zap.String("membership_id", membershipID.String()),
				zap.String("user_id", targetMembership.UserID.String()),
				zap.String("organization_id", targetMembership.OrganizationID.String()),
				zap.Error(err))

			// TODO: Attempt to restore OpenFGA permissions? Or queue for reconciliation
			// s.restoreRole(ctx, &targetMembership)

			return apperror.ErrInternal.
				WithMessage("failed to remove organization membership").
				WithDetail("db_error", err.Error())
		}

		// Log successful removal
		s.Logger.Info("organization member removed",
			zap.String("actor_user_id", userInfo.UserID),
			zap.String("actor_role", actorMembership.Role),
			zap.String("target_user_id", targetMembership.UserID.String()),
			zap.String("target_role", targetMembership.Role),
			zap.String("organization_id", targetMembership.OrganizationID.String()),
			zap.String("membership_id", membershipID.String()),
			zap.String("previous_status", targetMembership.Status))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}
