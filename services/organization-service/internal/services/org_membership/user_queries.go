package orgmembership

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/protoutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

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
func (s *OrgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.PaginationRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate pagination parameters
	if appErr := protoutils.ValidatePaginationReq(req); appErr != nil {
		return nil, appErr
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
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
	memberships, err := s.DBQ.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
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
func (s *OrgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 0. Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	// 1. Authenticate and extract user identity from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
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
	membership, err := s.DBQ.GetOrganizationMembership(ctx, membershipUUID)
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
