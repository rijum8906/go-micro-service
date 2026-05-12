package orgmembership

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

// LeaveOrganization removes the authenticated user from an organization.
//
// Execution Flow:
//   - Validate request parameters (ID and scoped token)
//   - Authenticate and extract user identity from context
//   - Validate scoped token
//   - Retrieve organization membership by ID
//   - Verify the authenticated user owns this membership
//   - Check if user has already left (idempotency)
//   - Prevent last owner from leaving (business rule)
//   - Update membership status to 'left' (soft delete)
//   - Remove user permissions from OpenFGA
//   - Return success response
//
// Idempotency:
//   - If membership status is already 'left', returns success immediately
//   - No error for duplicate leave requests
//   - OpenFGA permissions only removed once
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
//   - Returns PermissionDenied (not NotFound) for ownership mismatch
//   - Removes OpenFGA permissions immediately (defense in depth)
//
// Why check last owner?
//   - Prevents organizations from having zero owners
//   - Owners have special privileges and responsibilities
//   - Forces proper organization cleanup or ownership transfer
//
// Why soft delete (status='left') instead of hard delete?
//   - Maintains audit trail for compliance
//   - Allows reactivation if user rejoins
//   - Preserves historical data for analytics
//   - Prevents foreign key issues with related data
//
// Why remove OpenFGA permissions?
//   - User should lose all organization access immediately
//   - Prevents permission caching issues
//   - Defense in depth (database + authz)
//
// Error Responses:
//   - Validation: Invalid request, scoped token, or last owner cannot leave
//   - NotFound: Membership doesn't exist
//   - PermissionDenied: User doesn't own this membership
//   - Internal: Database or OpenFGA operation failed
//
// Example:
//
//	resp, err := service.LeaveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
//	    Id: membershipID,
//	    ScopedToken: token,
//	})
func (s *orgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDAndScopedTokenRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters (ID format and scoped token)
	if appErr := protoutils.ValidateIDAndScopedTokenReq(req); appErr != nil {
		return nil, appErr
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// 2. Validate token scope
	if req.TokenScope != string(token.TokenScopeLeaveOrganization) {
		return nil, apperror.ErrPermissionDenied.WithMessage("invalid token scope")
	}

	membership, err := s.q.GetOrganizationMembership(ctx, uuid.MustParse(req.Id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("membership not found")
		}
		return nil, apperror.ErrInternal.WithMessage("failed to fetch membership").WithDetail("db_error", err.Error())
	}

	// 3. Verify the authenticated user owns this membership
	// Prevents users from leaving other people's memberships (IDOR protection)
	if membership.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrPermissionDenied.WithMessage("you can only leave your own membership")
	}

	// 4. Idempotency check: Return success if already left
	if membership.Status == "left" {
		return &corev1.SuccessResponse{
			Success: true,
		}, nil
	}

	// 5. Business rule: Prevent last owner from leaving
	// Ensures organization always has at least one owner
	if membership.Role == "owner" {
		activeOwners, err := s.q.CountActiveOwnersByOrgID(ctx, membership.OrganizationID)
		if err != nil {
			return nil, apperror.ErrInternal.WithMessage("failed to check organization owners").WithDetail("db_error", err.Error())
		}

		if activeOwners <= 1 {
			return nil, apperror.ErrValidation.WithMessage("cannot leave organization: you are the last owner. Transfer ownership or delete the organization instead")
		}
	}

	// 6. Soft delete: Update membership status to 'left'
	_, err = s.q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
		ID:     membership.ID,
		Status: "left",
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to update membership").WithDetail("db_error", err.Error())
	}

	// 7. Remove user permissions from OpenFGA
	// Critical: User loses all organization access immediately
	if appErr := s.tuppleManager.Delete(ctx, []client.ClientTupleKeyWithoutCondition{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleMember,
			Object:   "organization:" + membership.OrganizationID.String(),
		},
	}); appErr != nil {
		// Note: Consider logging this error but not failing the operation?
		// The membership is already updated, OpenFGA failure leaves permission dangling
		return nil, appErr
	}

	// 8. Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// ############################ ORGANIZATION'S MEMBERSHIP MANAGEMENT ############################

// ChangeOrganizationMembershipStatus updates the status of an organization membership.
//
// This method handles all status transitions for organization memberships including leaving,
// banning, unbanning, and removal. It enforces role-based access control, validates status
// transitions, and maintains OpenFGA permission consistency.
//
// Execution Flow:
//  1. Validate request parameters (membership ID, token scope, target status)
//  2. Extract authenticated user identity from context
//  3. Retrieve target and actor memberships from database
//  4. Verify actor can manage target based on role hierarchy
//  5. Validate custom role permissions via OpenFGA
//  6. Update membership status in database
//  7. Synchronize OpenFGA permissions (remove on deactivation, restore on activation)
//  8. Return success response
//
// Status Transition Rules:
//   - active  → left:     User-initiated leave (no admin permission required)
//   - active  → removed:  Admin or owner action
//   - active  → banned:   Admin or owner action
//   - banned  → active:   Admin or owner action (unban)
//   - left    → active:   Not allowed (requires new invitation)
//   - removed → *:        Terminal state, no transitions allowed
//
// Security Constraints:
//   - Actors cannot modify their own status except to 'left'
//   - Owner status cannot be changed under any circumstances
//   - Last owner cannot be removed or have status changed
//   - Role hierarchy: owner > admin > member > custom roles
//
// Error Responses:
//   - Validation:      Invalid parameters, status transition, or token scope
//   - NotFound:        Target or actor membership does not exist
//   - PermissionDenied: Actor lacks required permissions for this operation
//   - Internal:        Database operation or OpenFGA synchronization failed
//
// TODO: Implement audit logging for compliance tracking
func (s *orgMembershipService) ChangeOrganizationMembershipStatus(
	ctx context.Context,
	req *org_membershipv1.ChangeOrgMembershipStatusReq,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	if appErr := validateChangeOrganizationStatusReq(req); appErr != nil {
		return nil, appErr
	}
	membershipID := uuid.MustParse(req.OrganizationMembershipId)

	// Extract authenticated user identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Load target and actor memberships
	membershipData, appErr := s.retrieveMemberships(ctx, membershipID, userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}
	actorMembership := membershipData.actor
	targetMembership := membershipData.target

	// Verify role-based access control (standard roles)
	if permissions.IsValidRole(actorMembership.Role) {
		if !permissions.CanActorManageTarget(actorMembership.Role, targetMembership.Role) {
			return nil, apperror.ErrPermissionDenied.WithMessage("you do not have permission to change this membership's status")
		}
	}

	// Verify custom role permissions via OpenFGA (handles cases standard RBAC misses)
	if appErr := utils.CheckCanChangeMembershipStatus(
		ctx, s.tuppleManager, actorMembership, targetMembership,
	); appErr != nil {
		return nil, appErr
	}

	// Persist status change to database
	if _, err := s.q.UpdateOrganizationMembershipStatus(ctx, db.UpdateOrganizationMembershipStatusParams{
		ID:     membershipID,
		Status: req.NewStatus,
	}); err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to update membership").
			WithDetail("db_error", err.Error())
	}

	// Synchronize OpenFGA permissions based on new status
	// - Non-active statuses (banned, left, removed): Remove all permissions
	// - Active status (unban): Restore previous permissions
	if req.NewStatus != constants.OrgMemStatusActive {
		s.removeRole(ctx, targetMembership)
	} else {
		s.addRole(ctx, targetMembership)
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *orgMembershipService) ChangeOrganizationMembershipRole(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipRoleReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) RemoveOrganizationMember(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
