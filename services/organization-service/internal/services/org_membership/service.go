package orgmembership

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

// ############################ USER QUERIES ############################

func (s *orgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.PaginationRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	// Step 1. Extract User Info
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("error", "failed to extract user info from context")
	}
	userID, _ := uuid.Parse(userInfo.UserID)

	// Step 2. Get Memberships
	memberships, err := s.q.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
		UserID: userID,
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get organization memberships from database").WithDetail("error", err.Error())
	}

	// Step 3. Map Memberships
	response := []*org_membershipv1.OrgMembershipRes{}
	for _, membership := range memberships {
		response = append(response, utils.MapOrgMembershipRes(&membership))
	}

	return &org_membershipv1.OrgMembershipsListRes{
		OrganizationMemberships: response,
	}, nil
}

func (s *orgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	membershipID, _ := uuid.Parse(req.Id)

	// Step 1. Get Membership
	membership, err := s.q.GetOrganizationMembership(ctx, membershipID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to get organization membership from database").WithDetail("error", err.Error())
	}

	// Step 2. Map Membership
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

// ############################ USER'S MEMBERSHIP MANAGEMENT ############################

func (s *orgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

// ############################ ORGANIZATION'S MEMBERSHIP MANAGEMENT ############################

func (s *orgMembershipService) CreateOrganizationMembership(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.OrganizationMembership, error) {
	return nil, nil
}

func (s *orgMembershipService) ChangeOrganizationMembershipStatus(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipStatusReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) ChangeOrganizationMembershipRole(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipRoleReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (s *orgMembershipService) RemoveOrganizationMember(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
