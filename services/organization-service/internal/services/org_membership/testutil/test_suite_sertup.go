package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/service_test_utils"
	orgmembership "github.com/rijum8906/relay/services/organization-service/internal/services/org_membership"
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
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("organization-service")))
	q := db.New(pool)
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient, &config.Env{
		InvitationTokenTTL: 7,
	})

	ctx := context.Background()

	t.Cleanup(func() {
		pool.Close()
	})

	permissionManager := permissions.NewPermissionManager(fgaClient)

	return &TestSuite{
		Pool:              pool,
		Q:                 q,
		TuppleManager:     coreopenfga.NewTupleManager(fgaClient),
		PermissionManager: permissionManager,
		Service:           service,
		Ctx:               ctx,
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
		to.Suite.Q.DeleteOrganizationMembershipHard(to.Suite.Ctx, membership.ID)
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
		to.Suite.Q.DeleteOrganizationMembershipHard(to.Suite.Ctx, membership.ID)
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
		to.Suite.Q.DeleteOrganizationMembershipHard(to.Suite.Ctx, membership.ID)
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
