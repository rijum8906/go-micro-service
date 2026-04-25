// Package organization service
package organization

import (
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

type OrganizationService struct {
	q db.Querier
}

func New(q db.Querier) organizationv1.OrganizationServiceServer {
	return &OrganizationService{
		q: q,
	}
}
