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
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

// ############################ USER QUERIES ############################

// GetMyMemberships retrieves all organization memberships for the authenticated user.
//
// Execution Flow:
//   - Extract User Identity from Context
//   - Validate pagination parameters
//   - Retrieve memberships from database by user ID
//   - Return memberships list with pagination
//
// Security:
//   - No OpenFGA check needed (users always have permission to view their own memberships)
//   - Uses authenticated user ID from context (cannot request other users' memberships)
//   - Returns ErrUnAuthenticated (not Internal) when user metadata missing
//
// Performance Optimizations:
//   - Pre-allocates response slice with capacity = len(memberships)
//   - Avoids multiple re-allocations during append operations
//   - Database query uses LIMIT and OFFSET for efficient pagination
//
// Pagination:
//   - Page numbers start at 1 (page 1 = offset 0)
//   - Offset calculated as: (page - 1) * limit
//   - Returns empty list if user has no memberships
//
// Why no permission check?
//   - Users always have implicit permission to view their own memberships
//   - The query is filtered by user ID from authenticated context
//   - No way to request another user's memberships (no user_id parameter)
//
// Error Responses:
//   - UnAuthenticated: User metadata not found in context
//   - Validation: Invalid pagination parameters (page < 1, limit out of range)
//   - Internal: Invalid user UUID in context or database query failed
//
// Use Cases:
//   - User dashboard showing all organizations they belong to
//   - Organization switcher dropdown in UI
//   - Profile page listing user's memberships
//
// Example Response:
//
//	{
//	  "organization_memberships": [
//	    {"organization_id": "org1", "role": "owner", "status": "active"},
//	    {"organization_id": "org2", "role": "member", "status": "active"}
//	  ]
//	}
func (s *orgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.PaginationRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate pagination parameters
	if appErr := protoutils.ValidatePaginationReq(req); appErr != nil {
		return nil, appErr
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	// 2. Parse user ID from context
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "invalid user uuid in context")
	}

	// 3. Calculate pagination offset
	// Page 1 = offset 0, Page 2 = offset limit, etc.
	offset := (req.Page - 1) * req.Limit

	// 4. Retrieve memberships from database filtered by user ID
	memberships, err := s.q.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
		UserID: userID,
		Limit:  req.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Transform and map response
	// Pre-allocating slice capacity improves performance by avoiding multiple re-allocations
	response := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		response = append(response, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: response,
	}, nil
}

// GetMyMembership retrieves the authenticated user's membership in a specific organization.
//
// Execution Flow:
//   - Authenticate and extract user identity from context
//   - Validate membership UUID format
//   - Retrieve membership from database
//   - Verify membership belongs to authenticated user
//   - Return membership details
//
// Security Features:
//   - Fetches membership by ID but implicitly validates ownership via UserID check
//   - Prevents IDOR (Insecure Direct Object Reference) attacks
//   - Database-level ownership validation (defense in depth)
//   - Returns permission denied (not not found) for mismatched ownership
//
// Why validate ownership at application level (not just database)?
//   - GetOrganizationMembership returns membership by ID regardless of UserID
//   - We need explicit ownership check to prevent users accessing others' memberships
//   - Error message uses PermissionDenied (not NotFound) to avoid information disclosure
//
// Why not use GetOrganizationMembershipByUserAndOrg?
//   - This method accepts membership ID (not organization ID)
//   - User might have multiple memberships across orgs
//   - Membership ID uniquely identifies the specific membership to retrieve
//
// Security Note:
//   - Returns ErrUnAuthenticated (not Internal) when user metadata missing
//   - Returns PermissionDenied (not NotFound) for ownership mismatch
//   - Prevents attackers from guessing valid membership IDs
//
// Error Responses:
//   - UnAuthenticated: User metadata not found in context
//   - Validation: Invalid membership UUID format
//   - NotFound: Membership doesn't exist
//   - PermissionDenied: Membership exists but belongs to different user
//   - Internal: Database query failed
//
// Use Case:
//   - User views their own membership details from a membership settings page
//   - User updates their own role or status (not implemented here)
func (s *orgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 0. Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	// 2. Validate membership UUID format
	membershipUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithDetail("id", "invalid membership uuid format")
	}

	// 3. Retrieve membership by ID
	// We fetch by ID only (not by ID + UserID) to allow explicit ownership check.
	// This gives us more control over error messaging (PermissionDenied vs NotFound).
	membership, err := s.q.GetOrganizationMembership(ctx, membershipUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithDetail("resource", "membership")
		}
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// 4. Security authorization check
	// Ensure the membership record actually belongs to the authenticated user.
	// This prevents IDOR attacks where users try to access others' memberships.
	if membership.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrPermissionDenied.WithDetail("reason", "membership ownership mismatch")
	}

	// 5. Return membership response
	return utils.MapOrgMembershipRes(&membership), nil
}

// ############################ ORGANIZATION QUERIES ############################

// GetOrganizationMembershipsByOrgID retrieves all members of an organization with pagination.
//
// Execution Flow:
//   - Extract User Identity from Context
//   - Check if organization exists
//   - Check if user has permission to view organization members
//   - Retrieve Memberships from Database with pagination
//   - Return Memberships list
//
// Why check organization existence before permission?
//   - Returns 404 immediately if organization doesn't exist
//   - Avoids unnecessary OpenFGA checks for non-existent orgs
//   - Better error clarity (NotFound vs PermissionDenied)
//
// Pagination Details:
//   - Page numbers start at 1 (page 1 = offset 0)
//   - Limit defines max items per page
//   - Offset calculated as: (page - 1) * limit
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//
// Error Responses:
//   - Validation: Invalid pagination params or organization UUID
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks view permission
//   - Internal: Database failure or missing user metadata
//
// Note: Returns empty list (not error) when organization has no members
func (s *orgMembershipService) GetOrganizationMembershipsByOrgID(ctx context.Context, req *corev1.IDWithPaginationReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate pagination parameters and organization ID
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	orgID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Verify organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error()) // FIXED: Added return
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check if user has permission to view organization members via OpenFGA
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.Id,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view members of this organization")
	}

	// 4. Retrieve organization memberships with pagination
	memberships, err := s.q.GetOrganizationMembershipsByOrgID(ctx, db.GetOrganizationMembershipsByOrgIDParams{
		OrganizationID: orgID,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit, // Page 1 = Offset 0
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Convert database models to proto response objects
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembershipsByRole retrieves organization memberships filtered by role (owner, member, admin, etc.).
//
// Execution Flow:
//   - Validate role and pagination parameters
//   - Authenticate and extract user identity from context
//   - Verify organization exists
//   - Check if user has permission to view organization members
//   - Retrieve memberships by organization ID and role with pagination
//   - Return filtered memberships list
//
// NOTE: Steps are identical to GetOrganizationMembershipsByOrgID except:
//   - Uses GetOrganizationMembershipsByOrgIDAndRole instead of GetOrganizationMembershipsByOrgID
//   - Adds role validation at the beginning
//   - Returns only members with the specified role
//
// Why check organization existence before permission?
//   - Early 404 if organization doesn't exist (better UX)
//   - Avoids unnecessary OpenFGA checks for non-existent orgs
//
// Valid Roles:
//   - "owner": Organization owners with full access
//   - "admin": Administrators with most permissions
//   - "member": Regular members with basic access
//   - "viewer": Read-only access (if applicable)
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//
// Pagination:
//   - Page numbers start at 1 (offset = (page-1) * limit)
//   - Empty list returned if no members with specified role
//
// Error Responses:
//   - Validation: Invalid role, pagination params, or organization ID format
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks view permission
//   - Internal: Database failure or missing user metadata
//
// Example Use Cases:
//   - "View all owners" page for sensitive settings
//   - "Member list" filtered by role in UI tabs
//   - Counting members by role for analytics
func (s *orgMembershipService) GetOrganizationMembershipsByRole(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByRoleReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate request parameters
	if !permissions.ValidateRole(req.Role) {
		return nil, apperror.ErrValidation.WithMessage("invalid role")
	}
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Verify organization exists
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error()) // FIXED: Added return
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check if user has permission to view organization members via OpenFGA
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view members of this organization")
	}

	// 4. Retrieve organization memberships filtered by role with pagination
	memberships, err := s.q.GetOrganizationMembershipsByOrgIDAndRole(ctx, db.GetOrganizationMembershipsByOrgIDAndRoleParams{
		OrganizationID: orgID,
		Role:           req.Role,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit, // Page 1 = Offset 0
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Convert database models to proto response objects
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembershipsByStatus retrieves organization memberships filtered by status (active, left, etc.).
//
// Execution Flow:
//   - Validate membership status and pagination parameters
//   - Authenticate and extract user identity from context
//   - Verify organization exists
//   - Check permission via OpenFGA (must have can_view_member)
//   - Retrieve memberships by organization ID and status with pagination
//   - Return filtered memberships list
//
// NOTE: Similar to GetOrganizationMembershipsByOrgID but adds status filtering:
//   - GetOrganizationMembershipsByOrgID: Returns ALL members (no status filter)
//   - GetOrganizationMembershipsByStatus: Returns ONLY members with specific status
//
// Why check organization existence before permission?
//   - Early validation: Returns 404 if org doesn't exist (better than permission denied)
//   - Avoids unnecessary OpenFGA checks for non-existent organizations
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//
// Valid Status Values:
//   - "active": Current members
//   - "left": Members who left the organization
//   - "removed": Members who were removed
//   - "invited": Pending invitations (if stored as memberships)
//
// Error Responses:
//   - Validation: Invalid status, pagination params, or organization ID format
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks view permission
//   - Internal: Database failure or missing user metadata
//
// Example Use Cases:
//   - Admin viewing only active members
//   - Audit log showing members who left
//   - Cleanup of removed members
func (s *orgMembershipService) GetOrganizationMembershipsByStatus(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByStatusReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate request parameters
	if !utils.ValidateOrgnaziationMembershipStatus(req.Status) {
		return nil, apperror.ErrValidation.WithMessage("invalid status")
	}
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid organization id")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Verify organization exists (returns 404 early if not found)
	exists, err := s.q.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error()) // Fixed: added return
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("organization not found")
	}

	// 3. Check permission via OpenFGA
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view members of this organization")
	}

	// 4. Retrieve memberships filtered by organization ID and status with pagination
	memberships, err := s.q.GetOrganizationMembershipsByOrgIDAndStatus(ctx, db.GetOrganizationMembershipsByOrgIDAndStatusParams{
		OrganizationID: orgID,
		Status:         req.Status,
		Limit:          req.Pagination.Limit,
		Offset:         (req.Pagination.Page - 1) * req.Pagination.Limit, // Page 1 = Offset 0
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 5. Convert database models to proto response objects
	result := []*org_membershipv1.OrgMembershipRes{}
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembership retrieves a single organization membership by ID.
//
// Execution Flow:
//   - Authenticate and extract user identity from context
//   - Validate membership UUID format
//   - Retrieve membership from database
//   - Check permission via OpenFGA
//   - Return membership details
//
// NOTE: Membership must be fetched BEFORE permission check because:
//   - Permission requires organization ID which is only available from the membership
//   - Early fetch provides immediate 404 if membership doesn't exist (better UX)
//   - Avoids unnecessary OpenFGA checks for non-existent memberships
//   - Permission check needs both user ID and organization ID to be valid
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//
// Error Responses:
//   - Validation: Nil request body or invalid membership UUID
//   - NotFound: Membership does not exist in database
//   - PermissionDenied: User lacks view permission for this organization
//   - Internal: Database query failed or user metadata missing
//
// Idempotency:
//   - Always returns same result for same membership ID (read-only operation)
func (s *orgMembershipService) GetOrganizationMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 0. Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("invalid request body")
	}

	orgMemID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid membership id")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("reason", "missing user metadata")
	}

	// 2. Retrieve membership from database
	membership, err := s.q.GetOrganizationMembership(ctx, orgMemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("membership not found")
		}
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch membership").WithDetail("db_error", err.Error())
	}

	// 3. Check permission via OpenFGA
	// User must have can_view_member permission on the organization
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMember,
		Object:   "organization:" + membership.OrganizationID.String(),
	})
	if appErr != nil {
		return nil, appErr
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.WithMessage("user does not have permission to view this organization")
	}

	// 4. Parse and return membership response
	return utils.MapOrgMembershipRes(&membership), nil
}

// ############################ INVITATION FLOW ############################

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
func (s *orgMembershipService) SendInvitation(ctx context.Context, req *org_membershipv1.SendInvitationRequest) (*corev1.SuccessResponse, error) {
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
	exists, err := s.userClient.CheckEmailExists(ctx, emailReq)
	if err != nil {
		return nil, apperror.ErrThirdParty.WithDetail("error", err.Error())
	}
	if !exists.Exists {
		return nil, apperror.ErrNotFound.WithMessage("email not found")
	}

	// 3. Check if sender has permission to add members via OpenFGA
	checkRes, appErr := s.tuppleManager.Check(ctx, client.ClientCheckRequest{
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
	membership, err := s.q.GetOrganizationMembershipByOrgIDAndUserID(ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
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
	tokenHash, appErr := s.hashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	// 6. Create invitation record in database
	_, err = s.q.CreateOrganizationInvitation(ctx, db.CreateOrganizationInvitationParams{
		Email:          req.Email,
		OrganizationID: uuid.MustParse(req.OrganizationId),
		Role:           req.Role,
		InvitedByMemID: membership.ID, // Uses membership ID from step 4
		TokenHash:      tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * time.Duration(s.config.InvitationTokenTTL)),
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
func (s *orgMembershipService) AcceptInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
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
	user, err := s.userClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 3. Retrieve invitation using the provided token hash
	invitation, err := s.q.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
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
	_, err = s.q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         uuid.MustParse(userInfo.UserID),
		OrganizationID: invitation.OrganizationID,
		Role:           invitation.Role,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create organization membership").WithDetail("db_error", err.Error())
	}

	// 7. Add user permissions to OpenFGA
	if appErr := s.tuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + userInfo.UserID,
			Relation: permissions.RoleMember,
			Object:   "organization:" + invitation.OrganizationID.String(),
		},
	}); appErr != nil {
		return nil, appErr
	}

	// 8. Mark the invitation as accepted
	_, err = s.q.AccecptOrganizationInvitation(ctx, db.AccecptOrganizationInvitationParams{
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
func (s *orgMembershipService) DeclineInvitation(ctx context.Context, req *corev1.TokenHashRequest) (*corev1.SuccessResponse, error) {
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
	user, err := s.userClient.GetUser(ctx, &corev1.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	// 3. Retrieve invitation using the provided token hash
	invitation, err := s.q.GetOrganizationInvitationByTokenHash(ctx, req.TokenHash)
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
	_, err = s.q.DeclineOrganizationInvitation(ctx, db.DeclineOrganizationInvitationParams{
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

// ############################ USER'S MEMBERSHIP MANAGEMENT ############################

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
