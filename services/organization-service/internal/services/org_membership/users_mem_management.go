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
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

// LeaveOrganization removes the authenticated user from an organization.
//
// Execution Flow:
//   - Validate request parameters (ID and scoped token)
//   - Authenticate and extract user identity from context
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

	// 2. Parse and retrieve organization membership by ID
	membershipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid membership id")
	}

	membership, err := s.q.GetOrganizationMembership(ctx, membershipID)
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

func (s *orgMembershipService) ChangeOrganizationMembershipStatus(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipStatusReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) ChangeOrganizationMembershipRole(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipRoleReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) RemoveOrganizationMember(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
