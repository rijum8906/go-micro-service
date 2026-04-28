// Package organization service
package organization

import (
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

type organizationService struct {
	q          db.Querier
	userClient userv1.UserServiceClient
}

func New(q db.Querier, client userv1.UserServiceClient) organizationv1.OrganizationServiceServer {
	return &organizationService{
		q:          q,
		userClient: client,
	}
}
