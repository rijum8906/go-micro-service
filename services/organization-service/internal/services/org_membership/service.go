package orgmembership

import (
	"context"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
)

// ############################ USER QUERIES ############################

func (m *orgMembershipService) GetMyMemberships(ctx context.Context, req *corev1.EmptyRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (m *orgMembershipService) GetMyMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	return nil, nil
}

// ############################ ORGANIZATION QUERIES ############################

func (m *orgMembershipService) GetOrganizationMembershipsByOrgID(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (m *orgMembershipService) GetOrganizationMembershipsByRole(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByRoleReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (m *orgMembershipService) GetOrganizationMembershipsByStatus(ctx context.Context, req *org_membershipv1.GetOrgMembershipsByStatusReq) (*org_membershipv1.OrgMembershipsListRes, error) {
	return nil, nil
}

func (m *orgMembershipService) GetOrganizationMembership(ctx context.Context, req *corev1.IDRequest) (*org_membershipv1.OrgMembershipRes, error) {
	return nil, nil
}

// ############################ USER'S MEMBERSHIP MANAGEMENT ############################

func (m *orgMembershipService) LeaveOrganization(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}

// ############################ ORGANIZATION'S MEMBERSHIP MANAGEMENT ############################

func (m *orgMembershipService) CreateOrganizationMembership(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.OrganizationMembership, error) {
	return nil, nil
}

func (m *orgMembershipService) ChangeOrganizationMembershipStatus(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipStatusReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (m *orgMembershipService) ChangeOrganizationMembershipRole(ctx context.Context, req *org_membershipv1.ChangeOrgMembershipRoleReq) (*corev1.SuccessResponse, error) {
	return nil, nil
}

func (m *orgMembershipService) RemoveOrganizationMember(ctx context.Context, req *corev1.IDRequest) (*corev1.SuccessResponse, error) {
	return nil, nil
}
