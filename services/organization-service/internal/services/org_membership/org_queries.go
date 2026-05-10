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
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

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
