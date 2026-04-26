// Package utils
package utils

import (
	"github.com/rijum8906/relay/packages/core/protoutils"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

func MapOrganization(org *db.Organization) *modelsv1.Organization {
	return &modelsv1.Organization{
		Id:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description.String,
		Status:      org.Status,
		LogoUrl:     org.LogoUrl.String,
		CreatedBy:   org.CreatedBy.String(),
		CreatedAt:   protoutils.MapTimestamp(org.CreatedAt),
		UpdatedAt:   protoutils.MapTimestamp(org.UpdatedAt),
		DeletedAt:   protoutils.MapTimestamp(org.DeletedAt),
		DeletedBy:   org.DeletedBy.String(),
	}
}
