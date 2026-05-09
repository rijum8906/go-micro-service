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

// ############################ USER QUERIES ############################

func (s *orgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.PaginationRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	// 0. Validate Pagination
	if appErr := protoutils.ValidatePaginationReq(req); appErr != nil {
		return nil, appErr
	}

	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "invalid user uuid in context")
	}

	offset := (req.Page - 1) * req.Limit

	// 2. Retrieve Memberships from Database
	memberships, err := s.q.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
		UserID: userID,
		Limit:  req.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to fetch memberships").WithDetail("db_error", err.Error())
	}

	// 4. Transform and Map Response
	// Pre-allocating the slice capacity improves performance by avoiding multiple re-allocations
	response := make([]*org_membershipv1.OrgMembershipRes, 0, len(memberships))
	for _, m := range memberships {
		response = append(response, utils.MapOrgMembershipRes(&m))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: response,
	}, nil
}

func (s *orgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	// 1. Authenticate and extract User Identity
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrUnAuthenticated.WithDetail("reason", "missing user metadata")
	}

	membershipUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithDetail("id", "invalid membership uuid format")
	}

	// 2. Retrieve Membership
	// We fetch by both ID and UserID to enforce ownership at the database level.
	// This prevents ID-guessing attacks (IDOR).
	membership, err := s.q.GetOrganizationMembership(ctx, membershipUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithDetail("resource", "membership")
		}
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	// 3. Security Authorization Check
	// Ensure the membership record actually belongs to the authenticated user.
	if membership.UserID.String() != userInfo.UserID {
		return nil, apperror.ErrPermissionDenied.WithDetail("reason", "membership ownership mismatch")
	}

	// 4. Response Mapping
	return utils.MapOrgMembershipRes(&membership), nil
}

// ############################ ORGANIZATION QUERIES ############################

func (s *orgMembershipService) GetOrganizationMembershipsByOrgID(ctx context.Context, req *corev1.IDWithPaginationReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (s *orgMembershipService) GetOrganizationMembershipsByRole(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByRoleReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (s *orgMembershipService) GetOrganizationMembershipsByStatus(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByStatusReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (s *orgMembershipService) GetOrganizationMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	return nil, nil
}

// ############################ INVITATION FLOW ############################

func (s *orgMembershipService) SendInvitation(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) AcceptInvitation(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) DeclineInvitation(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

// ############################ USER'S MEMBERSHIP MANAGEMENT ############################

func (s *orgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
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
