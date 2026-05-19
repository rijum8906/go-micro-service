// Package utils
package utils

import (
	"github.com/rijum8906/relay/packages/core/protoutils"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func MapOrganization(org *db.Organization) *modelsv1.Organization {
	return &modelsv1.Organization{
		Id:              org.ID.String(),
		Name:            org.Name,
		Slug:            org.Slug,
		Description:     org.Description.String,
		Status:          org.Status,
		LogoUrl:         org.LogoUrl.String,
		CreatedByUserId: org.CreatedByUserID.String(),
		CreatedAt:       protoutils.MapTimestamp(org.CreatedAt),
		UpdatedAt:       protoutils.MapTimestamp(org.UpdatedAt),
		DeletedAt:       protoutils.MapTimestamp(org.DeletedAt),
	}
}

func MapOrganizationInfo(org *db.Organization) *organizationv1.OrganizationResponse {
	return &organizationv1.OrganizationResponse{
		Id:     org.ID.String(),
		Name:   org.Name,
		Slug:   org.Slug,
		Status: org.Status,
	}
}

func MapOrgMembershipRes(membership *db.OrganizationMembership) *org_membershipv1.OrgMembershipRes {
	return &org_membershipv1.OrgMembershipRes{
		Id:             membership.ID.String(),
		UserId:         membership.UserID.String(),
		OrganizationId: membership.OrganizationID.String(),
		Role:           membership.Role,
		Status:         membership.Status,
		JoinedAt:       protoutils.MapTimestamp(membership.JoinedAt),
	}
}
