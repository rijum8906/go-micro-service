package orgmembership

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/protoutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

// SendInvitation sends an organization invitation to a user's email address.
//
// Execution Flow:
//   - Authenticate and extract user identity from context
//   - Validate request parameters (email, organization_id, role)
//   - Check if email is registered as a user in the system
//   - Verify sender has 'can_add_member' permission via OpenFGA
//   - Fetch sender's membership to get invited_by_mem_id
//   - Generate invitation token hash
//   - Create invitation record in database with 24-hour expiration
//   - Send invitation email to recipient (TODO: Not implemented yet)
//   - Return success response
//
// Why fetch membership before creating invitation?
//   - Invitation table requires 'invited_by_mem_id' (membership ID)
//   - The sender must have an active membership in the organization
//   - Membership fetch validates sender is actually a member
//   - Prevents invitations from non-members (defense in depth)
//
// Why check email existence before permission?
//   - Early validation fails fast if email doesn't exist
//   - Avoids unnecessary OpenFGA check for invalid emails
//   - Better user experience (clear error: "email not found" vs permission error)
//
// Permissions Required:
//   - can_add_member on organization:{organization_id}
//   - Sender must have active membership in the organization
//
// Configuration:
//   - Invitation TTL: s.config.InvitationTokenTTL (in days)
//   - Default expiration: 24 hours × TTL days
//
// Error Responses:
//   - Validation: Invalid email format, organization_id, or role
//   - NotFound: Email not registered in user service
//   - PermissionDenied: User lacks 'can_add_member' permission
//   - Internal: User service unavailable, database failure, or UUID parsing error
//
// TODO: Send invitation email via email service (currently missing)
// TODO: Add idempotency check for duplicate pending invitations
// TODO: Handle case where user already has a membership
func (s *OrgMembershipService) SendInvitation(ctx context.Context, req *org_membershipv1.SendInvitationRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters
	if appErr := protoutils.ValidateSendInvitationReq(req); appErr != nil {
		return nil, appErr
	}
	emailReq := &corev1.EmailRequest{
		Email: req.Email,
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}
	inviteBy, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("reason", "failed to parse user id").WithDetail("error", err.Error())
	}

	// 2. Check if email exists in user service (must be registered user)
	exists, err := s.UserClient.CheckEmailExists(ctx, emailReq)
	if err != nil {
		return nil, apperror.ErrThirdParty.WithDetail("error", err.Error())
	}
	if !exists.Exists {
		return nil, apperror.ErrNotFound.WithMessage("email not found")
	}

	// 3. Check if sender has permission to add members via OpenFGA
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + inviteBy.String(),
		Relation: permissions.PermissionCanAddMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to invite members")
	}

	// 4. Fetch sender's membership to get invited_by_mem_id
	// Required because invitation table references membership, not user directly
	membership, err := s.DBQ.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
		UserID:         inviteBy,
		OrganizationID: uuid.MustParse(req.OrganizationId),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("sender is not a member of this organization")
		}
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch membership").WithDetail("db_error", err.Error())
	}

	// 5. Generate secure token hash for invitation
	tokenHash, appErr := s.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	// 6. Create invitation record in database
	_, err = s.DBQ.CreateOrganizationInvitation(ctx, db.CreateOrganizationInvitationParams{
		Email:          req.Email,
		OrganizationID: uuid.MustParse(req.OrganizationId),
		Role:           req.Role,
		InvitedByMemID: membership.ID, // Uses membership ID from step 4
		TokenHash:      tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * time.Duration(s.Config.InvitationTokenTTL)),
			Valid: true,
		},
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to create invitation").WithDetail("db_error", err.Error())
	}

	// TODO: Send invitation email with token link
	// emailService.SendInvitationEmail(ctx, req.Email, tokenHash)

	// 7. Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// AcceptInvitation accepts a pending organization invitation and creates membership.
//
// Execution Flow:
//   - Authenticate and extract user identity from context
//   - Fetch user details from user service to get email for validation
//   - Retrieve invitation using the provided token hash
//   - Validate invitation is still valid (not expired, status = "pending")
//   - Verify authenticated user's email matches invitation recipient
//   - Create organization membership in database
//   - Add user to OpenFGA with member role
//   - Update invitation status to accepted
//   - Return success response
//
// NOTE: OpenFGA permission must be added AFTER membership creation because:
//   - Membership creation could fail (avoid orphaned OpenFGA permissions)
//   - OpenFGA requires organization ID which is available without membership ID
//   - Transaction consistency: If OpenFGA fails, membership can be rolled back
//   - Membership record provides audit trail even if OpenFGA fails later
//
// Idempotency:
//   - Already accepted invitations return not found (status != "pending")
//   - Each invitation can only be responded to once
//
// Permissions Required:
//   - User must be authenticated (user metadata in context)
//   - User email must match invitation recipient email
//   - No additional OpenFGA check needed (invitation token provides access)
//
// Error Responses:
//   - Validation: Nil request, empty token hash, or expired invitation
//   - NotFound: Invitation not found, already accepted, or already declined
//   - PermissionDenied: User email doesn't match invitation recipient
//   - Internal: User service unavailable, database failure, or invalid UUID format
func (s *OrgMembershipService) AcceptInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// 2. Fetch user details from user service to get email for validation
	user, err := s.UserClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 3. Retrieve invitation using the provided token hash
	invitation, err := s.DBQ.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("invitation not found or already expired")
		}
		return nil, apperror.ErrInternal.WithMessage("failed to fetch invitation").WithDetail("db_error", err.Error())
	}

	// 4. Validate invitation is still valid
	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.WithMessage("invitation has expired")
	}
	// Check if invitation is still in pending state (not accepted or cancelled)
	if invitation.Status != "pending" {
		return nil, apperror.ErrNotFound.WithMessage("invitation not found")
	}

	// 5. Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.WithMessage("this invitation was sent to a different email address")
	}

	// 6. Create organization membership for the user
	_, err = s.DBQ.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         uuid.MustParse(userInfo.UserID),
		OrganizationID: invitation.OrganizationID,
		Role:           invitation.Role,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create organization membership").WithDetail("db_error", err.Error())
	}

	// 7. Add user permissions to OpenFGA
	if appErr := s.TuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleMember,
			Object:   "organization:" + invitation.OrganizationID.String(),
		},
	}); appErr != nil {
		return nil, appErr
	}

	// 8. Mark the invitation as accepted
	_, err = s.DBQ.AccecptOrganizationInvitation(ctx, db.AccecptOrganizationInvitationParams{
		ID:          invitation.ID,
		RespondedBy: uuid.MustParse(userInfo.UserID),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to accept invitation").WithDetail("db_error", err.Error())
	}

	// 9. Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// DeclineInvitation rejects a pending organization invitation.
//
// Execution Flow:
//  1. Validate request parameters (non-nil, token hash not empty)
//  2. Extract authenticated user identity from context
//  3. Fetch user details from user service (email needed for validation)
//  4. Retrieve invitation by token hash from database
//  5. Validate invitation is still valid (not expired, status = "pending")
//  6. Verify user's email matches invitation recipient
//  7. Mark invitation as declined in database
//  8. Return success response
//
// Why fetch user from user service?
//   - Invitation contains email of intended recipient
//   - Need authenticated user's email to verify they are the intended recipient
//   - User service is the source of truth for user email addresses
//
// Idempotency:
//   - Already declined invitations return not found (status != "pending")
//   - Already accepted invitations return not found
//   - Each invitation can only be responded to once
//
// Permissions Required:
//   - User must be authenticated (user metadata in context)
//   - User email must match invitation recipient email
//   - No additional OpenFGA check needed (invitation token provides access)
//
// Error Responses:
//   - Validation: Nil request, empty token hash, or expired invitation
//   - NotFound: Invitation not found, already accepted, or already declined
//   - PermissionDenied: User email doesn't match invitation recipient
//   - Internal: User service unavailable, database failure, or invalid UUID format
//
// Notes:
//   - Does NOT create any organization membership (unlike AcceptInvitation)
//   - Does NOT add any OpenFGA permissions
//   - Only updates invitation status to 'declined'
//   - Sets RespondedBy and RespondedAt timestamps
func (s *OrgMembershipService) DeclineInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
	// 0. Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// 2. Fetch user details from user service to get email for validation
	user, err := s.UserClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 3. Retrieve invitation using the provided token hash
	invitation, err := s.DBQ.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("invitation not found or already expired")
		}
		return nil, apperror.ErrInternal.WithMessage("failed to fetch invitation").WithDetail("db_error", err.Error())
	}

	// 4. Validate invitation is still valid (pending, not expired)
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.WithMessage("invitation has expired")
	}
	if invitation.Status != "pending" {
		// Already accepted, declined, or cancelled
		return nil, apperror.ErrNotFound.WithMessage("invitation not found")
	}

	// 5. Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.WithMessage("this invitation was sent to a different email address")
	}

	// 6. Mark the invitation as declined
	_, err = s.DBQ.DeclineOrganizationInvitation(ctx, db.DeclineOrganizationInvitationParams{
		ID:          invitation.ID,
		RespondedBy: uuid.MustParse(userInfo.UserID),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to decline invitation").WithDetail("db_error", err.Error())
	}

	// 7. Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
