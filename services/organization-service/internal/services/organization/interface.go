// Package organization service
package organization

import (
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

type organizationService struct {
	q             db.Querier
	userClient    userv1.UserServiceClient
	tuppleManager coreopenfga.TuppleManager
}

func New(q db.Querier, client userv1.UserServiceClient, openFgaClient *coreopenfga.Client) organizationv1.OrganizationServiceServer {
	tuppleManager := coreopenfga.NewTupleManager(openFgaClient)
	return &organizationService{
		q:             q,
		userClient:    client,
		tuppleManager: tuppleManager,
	}
}
