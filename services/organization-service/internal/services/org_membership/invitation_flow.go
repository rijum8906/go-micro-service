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
	"github.com/rijum8906/relay/services/organization-service/app/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"go.uber.org/zap"
)

// SendInvitation sends an organization invitation to a user's email address.
//
// Execution Flow:
//   - Authenticate and extract user identity from context
//   - Validate request parameters (email, organization_id, role)
//   - Check if email is registered as a user in the system
//   - Get user ID from email for membership check
//   - Verify sender has 'can_add_member' permission via OpenFGA
//   - Fetch sender's membership to get invited_by_mem_id
//   - Validate inviter can assign the requested role
//   - Prevent self-invitation
//   - Generate invitation token hash
//   - Check for duplicate pending invitations (idempotency)
//   - Check if user is already a member
//   - Create invitation record in database with expiration
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
//   - Sender must have role hierarchy permission to assign target role
//
// Configuration:
//   - Invitation TTL: s.config.InvitationTokenTTL (in days)
//   - Default expiration: 24 hours × TTL days
//
// Error Responses:
//   - Validation: Invalid email format, organization_id, role, or self-invitation
//   - NotFound: Email not registered in user service
//   - PermissionDenied: User lacks 'can_add_member' permission or cannot assign role
//   - Conflict: User already has pending invitation or is already a member
//   - Internal: User service unavailable, database failure, or UUID parsing error
//
// TODO: Send invitation email via email service (currently missing)
func (s *OrgMembershipService) SendInvitation(
	ctx context.Context,
	req *org_membershipv1.SendInvitationRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	if appErr := protoutils.ValidateSendInvitationReq(req); appErr != nil {
		return nil, appErr
	}

	// Authenticate and extract user identity from context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	inviteBy, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("reason", "failed to parse user id").
			WithDetail("error", err.Error())
	}

	// Check if email exists in user service (must be registered user)
	emailReq := &corev1.EmailRequest{Email: req.Email}
	exists, err := s.UserClient.CheckEmailExists(ctx, emailReq)
	if err != nil {
		return nil, apperror.ErrThirdParty.
			WithMessage("failed to verify email existence").
			WithDetail("error", err.Error())
	}
	if !exists.Exists {
		return nil, apperror.ErrNotFound.
			WithMessage("email not registered").
			WithDetail("email", req.Email).
			WithDetail("reason", "User must have an account to receive invitations")
	}

	// Get user ID from email for membership check
	targetUser, err := s.UserClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, apperror.ErrThirdParty.
			WithMessage("failed to get user details").
			WithDetail("error", err.Error())
	}

	// Prevent self-invitation
	if inviteBy.String() == targetUser.Id {
		return nil, apperror.ErrValidation.
			WithMessage("cannot invite yourself to an organization").
			WithDetail("reason", "You are already a member or cannot invite yourself")
	}

	// Parse organization ID
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("invalid organization id").
			WithDetail("error", err.Error())
	}

	// Check if sender has permission to add members via OpenFGA
	// This runs outside transaction as it's an external check
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + inviteBy.String(),
		Relation: permissions.PermissionCanAddMember,
		Object:   "organization:" + orgID.String(),
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("user does not have permission to invite members").
			WithDetail("user_id", userInfo.UserID).
			WithDetail("organization_id", req.OrganizationId)
	}

	// Execute database operations in transaction
	var tokenHash string
	var invitedByMemID uuid.UUID

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Fetch sender's membership to get invited_by_mem_id
		membership, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
			UserID:         inviteBy,
			OrganizationID: orgID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperror.ErrNotFound.
					WithMessage("sender is not a member of this organization").
					WithDetail("user_id", userInfo.UserID).
					WithDetail("organization_id", req.OrganizationId)
			}
			return apperror.ErrInternal.
				WithMessage("failed to fetch sender membership").
				WithDetail("db_error", err.Error())
		}
		invitedByMemID = membership.ID

		// Validate inviter can assign the requested role
		if constants.IsStandardOrgRole(membership.Role) {
			if !permissions.CanActorManageTarget(membership.Role, req.Role) {
				return apperror.ErrPermissionDenied.
					WithMessage("insufficient permissions to invite user with this role").
					WithDetail("inviter_role", membership.Role).
					WithDetail("requested_role", req.Role).
					WithDetail("reason", "You cannot invite users with a role equal to or higher than yours")
			}
		}

		// Check for duplicate pending invitation (idempotency)
		existingInvitation, err := q.GetPendingInvitationByEmailAndOrg(ctx, db.GetPendingInvitationByEmailAndOrgParams{
			Email:          req.Email,
			OrganizationID: orgID,
		})
		if err == nil && existingInvitation.ExpiresAt.Time.After(time.Now()) {
			return apperror.ErrConflict.
				WithMessage("an active invitation already exists for this email").
				WithDetail("email", req.Email).
				WithDetail("organization_id", req.OrganizationId).
				WithDetail("expires_at", existingInvitation.ExpiresAt.Time.String())
		}

		// Check if user is already a member of the organization
		targetUserID, err := uuid.Parse(targetUser.Id)
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to parse target user id").
				WithDetail("error", err.Error())
		}

		existingMembership, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
			UserID:         targetUserID,
			OrganizationID: orgID,
		})
		if err == nil {
			if existingMembership.Status != constants.OrgMemStatusLeft &&
				existingMembership.Status != constants.OrgMemStatusRemoved {
				return apperror.ErrConflict.
					WithMessage("user is already a member of this organization").
					WithDetail("user_id", targetUser.Id).
					WithDetail("email", req.Email).
					WithDetail("current_status", existingMembership.Status)
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrInternal.
				WithMessage("failed to check existing membership").
				WithDetail("db_error", err.Error())
		}

		// Generate secure token hash for invitation
		hash, appErr := s.HashService.Generate(32)
		if appErr != nil {
			return appErr
		}
		tokenHash = hash

		// Create invitation record in database
		_, err = q.CreateOrganizationInvitation(ctx, db.CreateOrganizationInvitationParams{
			Email:          req.Email,
			OrganizationID: orgID,
			Role:           req.Role,
			InvitedByMemID: invitedByMemID,
			TokenHash:      tokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(time.Hour * 24 * time.Duration(s.Config.InvitationTokenTTL)),
				Valid: true,
			},
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to create invitation").
				WithDetail("db_error", err.Error())
		}

		s.Logger.Debug("invitation created in transaction",
			zap.String("email", req.Email),
			zap.String("organization_id", req.OrganizationId),
			zap.String("role", req.Role),
			zap.String("invited_by", userInfo.UserID))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// TODO: Send invitation email with token link (outside transaction)
	// Email service should be called asynchronously to not block the response
	// Consider using a background job queue (Redis, RabbitMQ, NATS)
	//
	// emailService.SendInvitationEmail(ctx, &email.SendInvitationRequest{
	//     Email:      req.Email,
	//     TokenHash:  tokenHash,
	//     OrgName:    orgName,
	//     InviterName: userInfo.Name,
	//     Role:       req.Role,
	//     ExpiresAt:  time.Now().Add(time.Hour * 24 * time.Duration(s.Config.InvitationTokenTTL)),
	// })
	//
	// NOTE: If email sending fails, the invitation is already created in DB
	// Implement a retry mechanism or background job for email sending

	s.Logger.Info("invitation created successfully",
		zap.String("email", req.Email),
		zap.String("organization_id", req.OrganizationId),
		zap.String("role", req.Role),
		zap.String("invited_by", userInfo.UserID),
		zap.String("token_hash_preview", tokenHash[:8])) // Log only first 8 chars for traceability

	// Return success response
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
//   - Begin database transaction
//   - Create organization membership in database
//   - Update invitation status to accepted
//   - Commit transaction
//   - Add user to OpenFGA with member role (outside transaction)
//   - Return success response
//
// Why OpenFGA after transaction?
//   - OpenFGA cannot be rolled back if membership creation fails
//   - Membership creation happens first (what we can rollback)
//   - If OpenFGA fails, membership exists but user lacks permissions (can retry)
//   - Alternative would be leaving orphaned OpenFGA entries (worse)
//
// Idempotency:
//   - Already accepted invitations return not found (status != "pending")
//   - Each invitation can only be responded to once
//   - Check for existing membership before creating
//
// Permissions Required:
//   - User must be authenticated (user metadata in context)
//   - User email must match invitation recipient email
//   - No additional OpenFGA check needed (invitation token provides access)
//
// Error Responses:
//   - Validation: Nil request, empty token hash, expired invitation
//   - NotFound: Invitation not found, already accepted, or already declined
//   - PermissionDenied: User email doesn't match invitation recipient
//   - Internal: User service unavailable, database failure, or invalid UUID format
//
// TODO: Implement retry mechanism for OpenFGA failures
// TODO: Add idempotency check for existing membership
func (s *OrgMembershipService) AcceptInvitation(
	ctx context.Context,
	req *corev1.TokenHashRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// Authenticate and extract user identity from context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Parse user ID early
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata contains invalid user id").
			WithDetail("error", err.Error())
	}

	// Fetch user details from user service to get email for validation
	user, err := s.UserClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, apperror.ErrThirdParty.
			WithMessage("failed to fetch user details").
			WithDetail("error", err.Error())
	}

	// Retrieve invitation using the provided token hash
	invitation, err := s.DBQ.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.
				WithMessage("invitation not found, already expired, or already processed")
		}
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch invitation").
			WithDetail("db_error", err.Error())
	}

	// Validate invitation is still valid
	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.
			WithMessage("invitation has expired").
			WithDetail("expired_at", invitation.ExpiresAt.Time.String()).
			WithDetail("reason", "Please request a new invitation")
	}

	// Check if invitation is still in pending state
	if invitation.Status != "pending" {
		return nil, apperror.ErrNotFound.
			WithMessage("invitation not found or already processed").
			WithDetail("current_status", invitation.Status).
			WithDetail("reason", "Invitation has already been accepted, declined, or revoked")
	}

	// Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("this invitation was sent to a different email address").
			WithDetail("invitation_email", invitation.Email).
			WithDetail("your_email", user.Email).
			WithDetail("reason", "You can only accept invitations sent to your email address")
	}

	// Parse organization ID
	orgID := invitation.OrganizationID

	// Execute database operations in transaction
	var membershipID uuid.UUID
	var membershipRole string

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// TODO: Add idempotency check for existing membership
		// Check if user is already a member (prev duplicate membership)
		existingMembership, err := q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
			UserID:         userID,
			OrganizationID: orgID,
		})
		if err == nil {
			if existingMembership.Status != constants.OrgMemStatusLeft &&
				existingMembership.Status != constants.OrgMemStatusRemoved {
				return apperror.ErrConflict.
					WithMessage("user is already a member of this organization").
					WithDetail("user_id", userInfo.UserID).
					WithDetail("organization_id", orgID.String()).
					WithDetail("current_status", existingMembership.Status).
					WithDetail("reason", "Cannot accept invitation for an existing membership")
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrInternal.
				WithMessage("failed to check existing membership").
				WithDetail("db_error", err.Error())
		}

		// Create organization membership for the user
		membership, err := q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
			UserID:         userID,
			OrganizationID: orgID,
			Role:           invitation.Role,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to create organization membership").
				WithDetail("db_error", err.Error())
		}
		membershipID = membership.ID
		membershipRole = membership.Role

		// Mark the invitation as accepted
		_, err = q.AcceptOrganizationInvitation(ctx, db.AcceptOrganizationInvitationParams{
			ID:          invitation.ID,
			RespondedBy: userID,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to accept invitation").
				WithDetail("db_error", err.Error())
		}

		s.Logger.Debug("invitation accepted and membership created in transaction",
			zap.String("invitation_id", invitation.ID.String()),
			zap.String("membership_id", membershipID.String()),
			zap.String("user_id", userInfo.UserID),
			zap.String("organization_id", orgID.String()),
			zap.String("role", invitation.Role))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Add user permissions to OpenFGA (outside transaction)
	// NOTE: If this fails, membership exists but user lacks permissions
	// This is preferable to the opposite (permissions without membership)
	if appErr := s.TuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleMember,
			Object:   "organization:" + orgID.String(),
		},
	}); appErr != nil {
		// CRITICAL: Membership created but OpenFGA sync failed
		// Log error and queue for retry
		s.Logger.Error("CRITICAL: Failed to add OpenFGA permissions after accepting invitation",
			zap.String("user_id", userInfo.UserID),
			zap.String("organization_id", orgID.String()),
			zap.String("membership_id", membershipID.String()),
			zap.String("role", membershipRole),
			zap.Error(appErr))

		// TODO: Queue for retry
		// s.enqueueOpenFGARetry(ctx, &OpenFGATask{
		//     Type:         "add_member_permissions",
		//     UserID:       userInfo.UserID,
		//     OrganizationID: orgID.String(),
		//     Role:         membershipRole,
		//     MembershipID: membershipID.String(),
		//     Attempt:      0,
		// })

		// Don't fail the request - user can still access with retry
		// Return success but note the issue in logs
	}

	// Log successful acceptance
	s.Logger.Info("user accepted organization invitation",
		zap.String("user_id", userInfo.UserID),
		zap.String("user_email", user.Email),
		zap.String("organization_id", orgID.String()),
		zap.String("role", invitation.Role),
		zap.String("invitation_id", invitation.ID.String()),
		zap.String("invited_by", invitation.InvitedByMemID.String()))

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

// DeclineInvitation rejects a pending organization invitation.
//
// Execution Flow:
//   - Validate request parameters (non-nil, token hash not empty)
//   - Extract authenticated user identity from context
//   - Fetch user details from user service (email needed for validation)
//   - Retrieve invitation by token hash from database
//   - Validate invitation is still valid (not expired, status = "pending")
//   - Verify user's email matches invitation recipient
//   - Begin database transaction
//   - Mark invitation as declined in database
//   - Commit transaction
//   - Return success response
//
// Why fetch user from user service?
//   - Invitation contains email of intended recipient
//   - Need authenticated user's email to verify they are the intended recipient
//   - User service is the source of truth for user email addresses
//
// Why use a transaction for decline?
//   - Only one database operation, but maintains consistency with other methods
//   - Future-proofing in case additional operations are added
//   - Consistent pattern with AcceptInvitation
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
func (s *OrgMembershipService) DeclineInvitation(
	ctx context.Context,
	req *corev1.TokenHashRequest,
) (*corev1.SuccessResponse, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}
	if req.TokenHash == "" {
		return nil, apperror.ErrValidation.WithMessage("token hash cannot be empty")
	}

	// Authenticate and extract user identity from context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("user metadata not found in context")
	}

	// Parse user ID
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata contains invalid user id").
			WithDetail("error", err.Error())
	}

	// Fetch user details from user service to get email for validation
	user, err := s.UserClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, apperror.ErrThirdParty.
			WithMessage("failed to fetch user details").
			WithDetail("error", err.Error())
	}

	// Retrieve invitation using the provided token hash
	invitation, err := s.DBQ.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.
				WithMessage("invitation not found, already expired, or already processed")
		}
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch invitation").
			WithDetail("db_error", err.Error())
	}

	// Validate invitation is still valid (pending, not expired)
	if time.Now().After(invitation.ExpiresAt.Time) {
		return nil, apperror.ErrValidation.
			WithMessage("invitation has expired").
			WithDetail("expired_at", invitation.ExpiresAt.Time.String()).
			WithDetail("reason", "Please request a new invitation")
	}

	if invitation.Status != "pending" {
		return nil, apperror.ErrNotFound.
			WithMessage("invitation not found or already processed").
			WithDetail("current_status", invitation.Status).
			WithDetail("reason", "Invitation has already been accepted, declined, or revoked")
	}

	// Verify the authenticated user's email matches the invitation recipient
	if invitation.Email != user.Email {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("this invitation was sent to a different email address").
			WithDetail("invitation_email", invitation.Email).
			WithDetail("your_email", user.Email).
			WithDetail("reason", "You can only decline invitations sent to your email address")
	}

	// Execute decline in transaction
	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Mark the invitation as declined
		_, err := q.DeclineOrganizationInvitation(ctx, db.DeclineOrganizationInvitationParams{
			ID:          invitation.ID,
			RespondedBy: userID,
		})
		if err != nil {
			return apperror.ErrInternal.
				WithMessage("failed to decline invitation").
				WithDetail("db_error", err.Error())
		}

		s.Logger.Debug("invitation declined in transaction",
			zap.String("invitation_id", invitation.ID.String()),
			zap.String("email", invitation.Email),
			zap.String("organization_id", invitation.OrganizationID.String()),
			zap.String("responded_by", userInfo.UserID))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Log successful decline
	s.Logger.Info("user declined organization invitation",
		zap.String("user_id", userInfo.UserID),
		zap.String("user_email", user.Email),
		zap.String("organization_id", invitation.OrganizationID.String()),
		zap.String("invitation_id", invitation.ID.String()),
		zap.String("invited_by", invitation.InvitedByMemID.String()))

	// Return success response
	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
