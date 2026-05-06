package orgmembership_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	orgmembership "github.com/rijum8906/relay/services/organization-service/internal/services/org_membership"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/services/testutils"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	// Load .env file before running tests
	if err := godotenv.Load("../../../.env"); err != nil {
		// Optional: fall back to .env.test
		if err := godotenv.Load("../../.env.test"); err != nil {
			// Skip if no .env file (for CI)
			if os.Getenv("CI") == "" {
				panic("No .env file found")
			}
		}
	}

	os.Exit(m.Run())
}

func Test_GetMyMemberships_Integration_Success(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	ownerID := uuid.New()

	org := servicetestutils.MustCreateOrg(ctx, q, &organizationv1.CreateOrganizationRequest{CreatedBy: ownerID.String()})
	ownerMembership, err := q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         ownerID,
		OrganizationID: org.ID,
		Role:           permissions.RoleOwner,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get the owner membership", func(t *testing.T) {
		// Attach the owner's user id in context
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		// Get memberships
		memberships, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(memberships.OrganizationMemberships) != 1 {
			t.Errorf("expected only 1 mebership record but got %d", len(memberships.OrganizationMemberships))
		}

		if memberships.OrganizationMemberships[0].Id != ownerMembership.ID.String() {
			t.Errorf("expected membership id to be %s but got %s", ownerMembership.ID.String(), memberships.OrganizationMemberships[0].Id)
		}
	})

	t.Run("Create and get admin memberships", func(t *testing.T) {
		// Create a new admin
		membership, err := q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
			UserID:         uuid.New(),
			Role:           permissions.RoleAdmin,
			OrganizationID: org.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Attach user info to context
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, membership.UserID.String(),
		))

		// Get my memberships
		memberships, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})
		if err != nil {
			t.Fatal(err)
		}

		if memberships.OrganizationMemberships[0].Id != membership.ID.String() {
			t.Errorf("expected membership id to be %s but got %s", membership.ID.String(), memberships.OrganizationMemberships[0].Id)
		}

		t.Cleanup(func() {
			q.DeleteOrganizationMembershipHard(ctx, membership.ID)
		})
	})

	t.Cleanup(func() {
		q.DeleteOrganizationMembershipHard(context.Background(), ownerMembership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}

func Test_GetMyMemberships_Integration_Failure(t *testing.T) {}
