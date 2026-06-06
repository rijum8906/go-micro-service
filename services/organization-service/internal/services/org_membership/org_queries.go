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
	"github.com/rijum8906/relay/services/organization-service/app/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
	"go.uber.org/zap"
)

// GetOrganizationMembershipsByOrgID retrieves all members of an organization with pagination.
//
// Execution Flow:
//   - Validate request parameters (pagination, organization ID)
//   - Extract authenticated user identity from context
//   - Check if organization exists (returns 404 if not found)
//   - Verify user has permission to view organization members via OpenFGA
//   - Retrieve memberships from database with pagination
//   - Convert and return memberships list
//
// Why check organization existence before permission?
//   - Returns 404 immediately if organization doesn't exist
//   - Avoids unnecessary OpenFGA checks for non-existent orgs
//   - Better error clarity (NotFound vs PermissionDenied)
//   - Prevents information disclosure (can't probe org existence via permission check)
//
// Why no transaction needed?
//   - This is a read-only operation
//   - No consistency guarantees needed across multiple reads
//   - Read-after-write consistency is handled by database default isolation level
//   - Transactions are for writes that need atomicity or isolation
//
// Pagination Details:
//   - Page numbers start at 1 (page 1 = offset 0)
//   - Limit defines max items per page (recommended: 10-100)
//   - Offset calculated as: (page - 1) * limit
//   - Returns empty list (not error) when organization has no members
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//   - User must be authenticated
//
// Error Responses:
//   - Validation: Invalid pagination params or organization UUID
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks 'can_view_member' permission
//   - Internal: Database failure or missing user metadata
//
// Example:
//
//	resp, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
//	    Id: "org-123",
//	    Pagination: &corev1.PaginationReq{Page: 1, Limit: 20},
//	})
func (s *OrgMembershipService) GetOrganizationMembershipsByOrgID(
	ctx context.Context,
	req *corev1.IDWithPaginationReq,
) (*org_membershipv1.OrgMembershipsListRes, error) {
	// Validate pagination parameters
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	// Validate and parse organization ID
	orgID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("invalid organization id").
			WithDetail("error", err.Error())
	}

	// Extract authenticated user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata not found in context").
			WithDetail("reason", "authentication required")
	}

	// Verify organization exists
	exists, err := s.DBQ.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to verify organization existence").
			WithDetail("db_error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.
			WithMessage("organization not found").
			WithDetail("organization_id", req.Id)
	}

	// Check if user has permission to view organization members via OpenFGA
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMembers,
		Object:   "organization:" + req.Id,
	})
	if appErr != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to check permissions").
			WithDetail("error", appErr.Error())
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("user does not have permission to view members of this organization").
			WithDetail("user_id", userInfo.UserID).
			WithDetail("organization_id", req.Id).
			WithDetail("required_permission", permissions.PermissionCanViewMembers)
	}

	// Calculate pagination offset (page 1 = offset 0)
	offset := (req.Pagination.Page - 1) * req.Pagination.Limit

	// Retrieve organization memberships with pagination
	memberships, err := s.DBQ.GetOrganizationMembershipsByOrgID(ctx, db.GetOrganizationMembershipsByOrgIDParams{
		OrganizationID: orgID,
		Limit:          req.Pagination.Limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch organization memberships").
			WithDetail("db_error", err.Error())
	}

	// Convert database models to proto response objects
	result := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	// Log for monitoring (optional, can be removed in production)
	s.Logger.Debug("retrieved organization memberships",
		zap.String("organization_id", req.Id),
		zap.String("user_id", userInfo.UserID),
		zap.Int64("page", int64(req.Pagination.Page)),
		zap.Int64("limit", int64(req.Pagination.Limit)),
		zap.Int("count", len(result)))

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembershipsByRole retrieves organization memberships filtered by role (owner, member, admin, etc.).
//
// Execution Flow:
//   - Validate role and pagination parameters
//   - Parse and validate organization ID
//   - Extract authenticated user identity from context
//   - Verify organization exists (returns 404 if not found)
//   - Check if user has permission to view organization members via OpenFGA
//   - Retrieve memberships by organization ID and role with pagination
//   - Convert and return filtered memberships list
//
// NOTE: Steps are identical to GetOrganizationMembershipsByOrgID except:
//   - Uses GetOrganizationMembershipsByOrgIDAndRole instead of GetOrganizationMembershipsByOrgID
//   - Adds role validation at the beginning
//   - Returns only members with the specified role
//
// Why check organization existence before permission?
//   - Returns 404 immediately if organization doesn't exist
//   - Avoids unnecessary OpenFGA checks for non-existent orgs
//   - Better error clarity (NotFound vs PermissionDenied)
//   - Prevents information disclosure (can't probe org existence via permission check)
//
// Why no transaction needed?
//   - This is a read-only operation
//   - No consistency guarantees needed across multiple reads
//   - Database default isolation level is sufficient
//
// Valid Roles:
//   - "owner": Organization owners with full access
//   - "admin": Administrators with most permissions
//   - "member": Regular members with basic access
//   - "viewer": Read-only access (if applicable)
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//   - User must be authenticated
//
// Pagination:
//   - Page numbers start at 1 (page 1 = offset 0)
//   - Limit defines max items per page (recommended: 10-100)
//   - Offset calculated as: (page - 1) * limit
//   - Empty list returned if no members with specified role
//
// Error Responses:
//   - Validation: Invalid role, pagination params, or organization ID format
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks 'can_view_member' permission
//   - Internal: Database failure or missing user metadata
//
// Example Use Cases:
//   - "View all owners" page for sensitive settings
//   - "Member list" filtered by role in UI tabs
//   - Counting members by role for analytics
//
// Example:
//
//	resp, err := service.GetOrganizationMembershipsByRole(ctx, &org_membershipv1.GetOrgMembershipsByRoleReq{
//	    OrganizationId: "org-123",
//	    Role: "admin",
//	    Pagination: &corev1.PaginationReq{Page: 1, Limit: 20},
//	})
func (s *OrgMembershipService) GetOrganizationMembershipsByRole(
	ctx context.Context,
	req *org_membershipv1.GetOrgMembershipsByRoleReq,
) (*org_membershipv1.OrgMembershipsListRes, error) {
	// Validate role parameter
	if !permissions.IsValidRole(req.Role) {
		return nil, apperror.ErrValidation.
			WithMessage("invalid role")
	}

	// Validate pagination parameters
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	// Parse and validate organization ID
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("invalid organization id").
			WithDetail("error", err.Error())
	}

	// Extract authenticated user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata not found in context").
			WithDetail("reason", "authentication required")
	}

	// Verify organization exists
	exists, err := s.DBQ.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to verify organization existence").
			WithDetail("db_error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.
			WithMessage("organization not found").
			WithDetail("organization_id", req.OrganizationId)
	}

	// Check if user has permission to view organization members via OpenFGA
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMembers,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to check permissions").
			WithDetail("error", appErr.Error())
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("user does not have permission to view members of this organization").
			WithDetail("user_id", userInfo.UserID).
			WithDetail("organization_id", req.OrganizationId).
			WithDetail("required_permission", permissions.PermissionCanViewMembers)
	}

	// Calculate pagination offset (page 1 = offset 0)
	offset := (req.Pagination.Page - 1) * req.Pagination.Limit

	// Retrieve organization memberships filtered by role with pagination
	memberships, err := s.DBQ.GetOrganizationMembershipsByOrgIDAndRole(ctx, db.GetOrganizationMembershipsByOrgIDAndRoleParams{
		OrganizationID: orgID,
		Role:           req.Role,
		Limit:          req.Pagination.Limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch organization memberships by role").
			WithDetail("role", req.Role).
			WithDetail("db_error", err.Error())
	}

	// Convert database models to proto response objects
	result := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	// Log for monitoring (optional)
	s.Logger.Debug("retrieved organization memberships by role",
		zap.String("organization_id", req.OrganizationId),
		zap.String("role", req.Role),
		zap.String("user_id", userInfo.UserID),
		zap.Int32("page", req.Pagination.Page),
		zap.Int32("limit", req.Pagination.Limit),
		zap.Int("count", len(result)))

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembershipsByStatus retrieves organization memberships filtered by status (active, left, etc.).
//
// Execution Flow:
//   - Validate membership status and pagination parameters
//   - Parse and validate organization ID
//   - Extract authenticated user identity from context
//   - Verify organization exists (returns 404 if not found)
//   - Check permission via OpenFGA (must have can_view_member)
//   - Retrieve memberships by organization ID and status with pagination
//   - Convert and return filtered memberships list
//
// NOTE: Similar to GetOrganizationMembershipsByOrgID but adds status filtering:
//   - GetOrganizationMembershipsByOrgID: Returns ALL active members (status NOT IN ('left', 'removed'))
//   - GetOrganizationMembershipsByStatus: Returns members with SPECIFIC status (including left/removed)
//
// Why check organization existence before permission?
//   - Returns 404 immediately if organization doesn't exist
//   - Avoids unnecessary OpenFGA checks for non-existent orgs
//   - Better error clarity (NotFound vs PermissionDenied)
//   - Prevents information disclosure (can't probe org existence via permission check)
//
// Why no transaction needed?
//   - This is a read-only operation
//   - No consistency guarantees needed across multiple reads
//   - Database default isolation level is sufficient
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//   - User must be authenticated
//
// Valid Status Values:
//   - "active": Current active members (has permissions)
//   - "banned": Banned members (no access, can be unbanned)
//   - "suspended": Temporarily suspended members
//   - "left": Members who left voluntarily
//   - "removed": Members who were removed (terminal state)
//
// Pagination:
//   - Page numbers start at 1 (page 1 = offset 0)
//   - Limit defines max items per page (recommended: 10-100)
//   - Offset calculated as: (page - 1) * limit
//   - Empty list returned if no members with specified status
//
// Error Responses:
//   - Validation: Invalid status, pagination params, or organization ID format
//   - NotFound: Organization doesn't exist
//   - PermissionDenied: User lacks 'can_view_member' permission
//   - Internal: Database failure or missing user metadata
//
// Example Use Cases:
//   - Admin viewing only banned members for review
//   - Audit log showing members who left
//   - Cleanup of removed members
//   - Monitoring suspended members
//
// Example:
//
//	resp, err := service.GetOrganizationMembershipsByStatus(ctx, &org_membershipv1.GetOrgMembershipsByStatusReq{
//	    OrganizationId: "org-123",
//	    Status: "banned",
//	    Pagination: &corev1.PaginationReq{Page: 1, Limit: 20},
//	})
func (s *OrgMembershipService) GetOrganizationMembershipsByStatus(
	ctx context.Context,
	req *org_membershipv1.GetOrgMembershipsByStatusReq,
) (*org_membershipv1.OrgMembershipsListRes, error) {
	// Validate status parameter
	if !constants.IsValidaOrgMemStatus(req.Status) {
		return nil, apperror.ErrValidation.
			WithMessage("invalid membership status")
	}

	// Validate pagination parameters
	if appErr := protoutils.ValidatePaginationReq(req.Pagination); appErr != nil {
		return nil, appErr
	}

	// Parse and validate organization ID
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("invalid organization id").
			WithDetail("error", err.Error())
	}

	// Extract authenticated user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata not found in context").
			WithDetail("reason", "authentication required")
	}

	// Verify organization exists (returns 404 early if not found)
	exists, err := s.DBQ.CheckOrganizationExists(ctx, orgID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to verify organization existence").
			WithDetail("db_error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrNotFound.
			WithMessage("organization not found").
			WithDetail("organization_id", req.OrganizationId)
	}

	// Check permission via OpenFGA
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMembers,
		Object:   "organization:" + req.OrganizationId,
	})
	if appErr != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to check permissions").
			WithDetail("error", appErr.Error())
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("user does not have permission to view members of this organization").
			WithDetail("user_id", userInfo.UserID).
			WithDetail("organization_id", req.OrganizationId).
			WithDetail("required_permission", permissions.PermissionCanViewMembers)
	}

	// Calculate pagination offset (page 1 = offset 0)
	offset := (req.Pagination.Page - 1) * req.Pagination.Limit

	// Retrieve memberships filtered by organization ID and status with pagination
	memberships, err := s.DBQ.GetOrganizationMembershipsByOrgIDAndStatus(ctx, db.GetOrganizationMembershipsByOrgIDAndStatusParams{
		OrganizationID: orgID,
		Status:         req.Status,
		Limit:          req.Pagination.Limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch organization memberships by status").
			WithDetail("status", req.Status).
			WithDetail("db_error", err.Error())
	}

	// Convert database models to proto response objects
	result := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		result = append(result, utils.MapOrgMembershipRes(&m))
	}

	// Log for monitoring (optional)
	s.Logger.Debug("retrieved organization memberships by status",
		zap.String("organization_id", req.OrganizationId),
		zap.String("status", req.Status),
		zap.String("user_id", userInfo.UserID),
		zap.Int32("page", req.Pagination.Page),
		zap.Int32("limit", req.Pagination.Limit),
		zap.Int("count", len(result)))

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: result,
	}, nil
}

// GetOrganizationMembership retrieves a single organization membership by ID.
//
// Execution Flow:
//   - Validate request parameters (non-nil, valid UUID)
//   - Extract authenticated user identity from context
//   - Retrieve membership from database
//   - Check permission via OpenFGA (requires organization ID from membership)
//   - Return membership details
//
// NOTE: Membership must be fetched BEFORE permission check because:
//   - Permission requires organization ID which is only available from the membership
//   - Early fetch provides immediate 404 if membership doesn't exist (better UX)
//   - Avoids unnecessary OpenFGA checks for non-existent memberships
//   - Permission check needs both user ID and organization ID to be valid
//
// Why no transaction needed?
//   - This is a read-only operation
//   - No consistency guarantees needed across multiple reads
//   - Single database query with default isolation level is sufficient
//
// Permissions Required:
//   - can_view_member on organization:{organization_id}
//   - User must be authenticated
//
// Error Responses:
//   - Validation: Nil request body or invalid membership UUID format
//   - NotFound: Membership does not exist in database
//   - PermissionDenied: User lacks 'can_view_member' permission for this organization
//   - Internal: Database query failed or user metadata missing
//
// Idempotency:
//   - Always returns same result for same membership ID (read-only operation)
//   - Safe to retry indefinitely
//
// Example:
//
//	resp, err := service.GetOrganizationMembership(ctx, &corev1.IDRequest{
//	    Id: membershipID,
//	})
func (s *OrgMembershipService) GetOrganizationMembership(
	ctx context.Context,
	req *corev1.IDRequest,
) (*org_membershipv1.OrgMembershipRes, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.
			WithMessage("request body cannot be nil")
	}

	// Parse and validate membership ID
	orgMemID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.
			WithMessage("invalid membership id").
			WithDetail("error", err.Error())
	}

	// Extract authenticated user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithMessage("user metadata not found in context").
			WithDetail("reason", "authentication required")
	}

	// Retrieve membership from database
	membership, err := s.DBQ.GetOrganizationMembership(ctx, orgMemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.
				WithMessage("membership not found").
				WithDetail("membership_id", req.Id)
		}
		return nil, apperror.ErrInternal.
			WithMessage("failed to fetch organization membership").
			WithDetail("db_error", err.Error())
	}

	// Check permission via OpenFGA
	// User must have can_view_member permission on the organization
	checkRes, appErr := s.TuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + userInfo.UserID,
		Relation: permissions.PermissionCanViewMembers,
		Object:   "organization:" + membership.OrganizationID.String(),
	})
	if appErr != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to check permissions").
			WithDetail("error", appErr.Error())
	}
	if !*checkRes.Allowed {
		return nil, apperror.ErrPermissionDenied.
			WithMessage("user does not have permission to view this organization").
			WithDetail("user_id", userInfo.UserID).
			WithDetail("organization_id", membership.OrganizationID.String()).
			WithDetail("required_permission", permissions.PermissionCanViewMembers)
	}

	// Log for monitoring (optional)
	s.Logger.Debug("retrieved organization membership",
		zap.String("membership_id", req.Id),
		zap.String("organization_id", membership.OrganizationID.String()),
		zap.String("user_id", userInfo.UserID),
		zap.String("role", membership.Role),
		zap.String("status", membership.Status))

	// Convert and return membership response
	return utils.MapOrgMembershipRes(&membership), nil
}
