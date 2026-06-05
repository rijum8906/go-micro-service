package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/repositories/openfga"
	"github.com/rijum8906/relay/services/organization-service/internal/services/helper"
	orgmembership "github.com/rijum8906/relay/services/organization-service/internal/services/org_membership"
	"go.uber.org/zap"
)

type TestSuite struct {
	Pool              *pgxpool.Pool
	Q                 db.Querier
	TuppleManager     coreopenfga.TuppleManager
	PermissionManager *permissions.PermissionManager
	Service           org_membershipv1.OrganizationMembershipServiceServer
	Ctx               context.Context
}

func NewTestSuite(t *testing.T, fgaClient *coreopenfga.Client) *TestSuite {
	service := orgmembership.NewTestService(fgaClient)
	service.Helper, service.OpenFGARepo = newTestServiceDependencies(
		service.DBPool,
		service.DBQ,
		fgaClient,
		service.Logger,
	)

	ctx := context.Background()

	t.Cleanup(func() {
		service.DBPool.Close()
	})

	permissionManager := permissions.NewPermissionManager(fgaClient)

	return &TestSuite{
		Pool:              service.DBPool,
		Q:                 service.DBQ,
		TuppleManager:     coreopenfga.NewTupleManager(fgaClient),
		PermissionManager: permissionManager,
		Service:           service,
		Ctx:               ctx,
	}
}

// noopPublisher keeps membership-management tests focused on database state and
// permission decisions. Production code still calls the repository publish
// methods, but tests do not need a NATS server to verify these service flows.
type noopPublisher struct{}

func (noopPublisher) Publish(string, any) *apperror.AppError {
	return nil
}

func (noopPublisher) PublishAsync(string, any) (nats.PubAckFuture, *apperror.AppError) {
	return nil, nil
}

func (noopPublisher) PublishWithHeaders(string, any, nats.Header) *apperror.AppError {
	return nil
}

func newTestServiceDependencies(
	dbPool *pgxpool.Pool,
	dbq *db.Queries,
	fgaClient *coreopenfga.Client,
	logger *zap.Logger,
) (*helper.ServiceHelper, *openfga.Repository) {
	publisher := noopPublisher{}
	tupleManager := coreopenfga.NewTupleManager(fgaClient)
	permissionManager := permissions.NewPermissionManager(fgaClient)

	return &helper.ServiceHelper{
			DBPool:              dbPool,
			DBQ:                 dbq,
			TuppleManager:       tupleManager,
			PermissionManager:   permissionManager,
			Logger:              logger,
			OrgOpenFGAPublisher: publisher,
		}, &openfga.Repository{
			DBPool:              dbPool,
			DBQ:                 dbq,
			TuppleManager:       tupleManager,
			PermissionManager:   permissionManager,
			Logger:              logger,
			OrgOpenFGAPublisher: publisher,
		}
}

type OrgTestSuite struct {
	Suite            *TestSuite
	Org              *db.Organization
	OwnerMembership  *db.OrganizationMembership
	AdminMembership  *db.OrganizationMembership
	MemberMembership *db.OrganizationMembership
}

func (ts *TestSuite) CreateOrg(t *testing.T, ownerID uuid.UUID) *OrgTestSuite {
	org, err := ts.Q.CreateOrganization(ts.Ctx, db.CreateOrganizationParams{
		Name:        testutils.GenerateRandomString(5),
		Slug:        strings.ToLower(testutils.GenerateRandomString(10)),
		Description: pgtype.Text{String: testutils.GenerateRandomString(20), Valid: true},
		CreatedBy:   ownerID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ts.Q.DeleteOrganizationHard(ts.Ctx, org.ID)
	})

	return &OrgTestSuite{
		Suite: ts,
		Org:   &org,
	}
}

func (to *OrgTestSuite) CreateOwner(t *testing.T, ownerID uuid.UUID) {
	membership, err := to.Suite.Q.CreateOrganizationMembershipOwner(to.Suite.Ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         ownerID,
		OrganizationID: to.Org.ID,
		Role:           permissions.RoleOwner,
	})
	if err != nil {
		t.Fatal(err)
	}

	if appErr := to.Suite.TuppleManager.Write(to.Suite.Ctx, []client.ClientTupleKey{
		{
			User:     "user:" + ownerID.String(),
			Relation: permissions.RoleOwner,
			Object:   "organization:" + to.Org.ID.String(),
		},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	t.Cleanup(func() {
		to.Suite.Q.HardDeleteOrganizationMembership(to.Suite.Ctx, membership.ID)
		to.Suite.TuppleManager.Delete(to.Suite.Ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + ownerID.String(),
				Relation: permissions.RoleOwner,
				Object:   "organization:" + to.Org.ID.String(),
			},
		})
	})

	to.OwnerMembership = &membership
}

func (to *OrgTestSuite) CreateAdmin(t *testing.T, adminID uuid.UUID) {
	membership, err := to.Suite.Q.CreateOrganizationMembership(to.Suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         adminID,
		OrganizationID: to.Org.ID,
		Role:           permissions.RoleAdmin,
	})
	if err != nil {
		t.Fatal(err)
	}

	if appErr := to.Suite.TuppleManager.Write(to.Suite.Ctx, []client.ClientTupleKey{
		{
			User:     "user:" + adminID.String(),
			Relation: permissions.RoleAdmin,
			Object:   "organization:" + to.Org.ID.String(),
		},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	t.Cleanup(func() {
		to.Suite.Q.HardDeleteOrganizationMembership(to.Suite.Ctx, membership.ID)
		to.Suite.TuppleManager.Delete(to.Suite.Ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + adminID.String(),
				Relation: permissions.RoleAdmin,
				Object:   "organization:" + to.Org.ID.String(),
			},
		})
	})

	to.AdminMembership = &membership
}

func (to *OrgTestSuite) CreateMember(t *testing.T, memberID uuid.UUID) {
	membership, err := to.Suite.Q.CreateOrganizationMembership(to.Suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         memberID,
		OrganizationID: to.Org.ID,
		Role:           permissions.RoleMember,
	})
	if err != nil {
		t.Fatal(err)
	}

	if appErr := to.Suite.TuppleManager.Write(to.Suite.Ctx, []client.ClientTupleKey{
		{
			User:     "user:" + memberID.String(),
			Relation: permissions.RoleMember,
			Object:   "organization:" + to.Org.ID.String(),
		},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	t.Cleanup(func() {
		to.Suite.Q.HardDeleteOrganizationMembership(to.Suite.Ctx, membership.ID)
		to.Suite.TuppleManager.Delete(to.Suite.Ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + memberID.String(),
				Relation: permissions.RoleMember,
				Object:   "organization:" + to.Org.ID.String(),
			},
		})
	})

	to.MemberMembership = &membership
}
