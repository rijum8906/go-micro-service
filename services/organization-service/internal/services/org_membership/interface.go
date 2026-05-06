// Package orgmembership
package orgmembership

import (
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

type orgMembershipService struct {
	q             db.Querier
	userClient    userv1.UserServiceClient
	tuppleManager coreopenfga.TuppleManager
}

func New(q db.Querier, client userv1.UserServiceClient, openFgaClient *coreopenfga.Client) org_membershipv1.OrganizationMembershipServiceServer {
	tuppleManager := coreopenfga.NewTupleManager(openFgaClient)
	return &orgMembershipService{
		q:             q,
		userClient:    client,
		tuppleManager: tuppleManager,
	}
}
