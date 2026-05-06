package orgmembership_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
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

func Test_GetMyMemberships_Integration_Failure(t *testing.T) {
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

	t.Run("Without user metadata", func(t *testing.T) {
		_, err := service.GetMyMemberships(context.Background(), &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeUnAuthenticated)) {
			t.Errorf("expected unauthenticated error, got %s", err.Error())
		}
	})

	t.Run("With invalid pagination", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  0,
			Limit: 50,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected validation error, got %s", err.Error())
		}
	})

	t.Run("With invalid user id in metadata", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, "invalid-user-id",
		))

		_, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeInternal)) {
			t.Errorf("expected internal error, got %s", err.Error())
		}
	})

	t.Cleanup(func() {
		q.DeleteOrganizationMembershipHard(context.Background(), ownerMembership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}

func Test_GetMyMembership_Integration_Success(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	userID := uuid.New()
	org := servicetestutils.MustCreateOrg(ctx, q, &organizationv1.CreateOrganizationRequest{CreatedBy: userID.String()})
	membership, err := q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         userID,
		OrganizationID: org.ID,
		Role:           permissions.RoleOwner,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		dto.MetaUserIDKey, userID.String(),
	))

	res, err := service.GetMyMembership(ctx, &corev1.IDRequest{
		Id: membership.ID.String(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.Id != membership.ID.String() {
		t.Errorf("expected membership id to be %s but got %s", membership.ID.String(), res.Id)
	}
	if res.UserId != userID.String() {
		t.Errorf("expected user id to be %s but got %s", userID.String(), res.UserId)
	}
	if res.OrganizationId != org.ID.String() {
		t.Errorf("expected organization id to be %s but got %s", org.ID.String(), res.OrganizationId)
	}

	t.Cleanup(func() {
		q.DeleteOrganizationMembershipHard(context.Background(), membership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}

func Test_GetMyMembership_Integration_Failure(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	ownerID := uuid.New()
	otherUserID := uuid.New()
	org := servicetestutils.MustCreateOrg(ctx, q, &organizationv1.CreateOrganizationRequest{CreatedBy: ownerID.String()})
	ownerMembership, err := q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         ownerID,
		OrganizationID: org.ID,
		Role:           permissions.RoleOwner,
	})
	if err != nil {
		t.Fatal(err)
	}
	otherMembership, err := q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         otherUserID,
		OrganizationID: org.ID,
		Role:           permissions.RoleMember,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Without user metadata", func(t *testing.T) {
		_, err := service.GetMyMembership(context.Background(), &corev1.IDRequest{
			Id: ownerMembership.ID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeUnAuthenticated)) {
			t.Errorf("expected unauthenticated error, got %s", err.Error())
		}
	})

	t.Run("With invalid membership id", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetMyMembership(ctx, &corev1.IDRequest{
			Id: "invalid-membership-id",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected validation error, got %s", err.Error())
		}
	})

	t.Run("With unknown membership id", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetMyMembership(ctx, &corev1.IDRequest{
			Id: uuid.NewString(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeNotFound)) {
			t.Errorf("expected not found error, got %s", err.Error())
		}
	})

	t.Run("With ownership mismatch", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetMyMembership(ctx, &corev1.IDRequest{
			Id: otherMembership.ID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodePermissionDenied)) {
			t.Errorf("expected permission denied error, got %s", err.Error())
		}
	})

	t.Cleanup(func() {
		q.DeleteOrganizationMembershipHard(context.Background(), otherMembership.ID)
		q.DeleteOrganizationMembershipHard(context.Background(), ownerMembership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}

func Test_GetOrganizationMembershipsByOrgID_Integration_Success(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)

	ctx := context.Background()

	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()

	org := servicetestutils.MustCreateOrg(ctx, q, &organizationv1.CreateOrganizationRequest{CreatedBy: ownerID.String()})
	ownerMembership, err := q.CreateOrganizationMembershipOwner(ctx, db.CreateOrganizationMembershipOwnerParams{
		UserID:         ownerID,
		OrganizationID: org.ID,
		Role:           permissions.RoleOwner,
	})
	if err != nil {
		t.Fatal(err)
	}
	adminMembership, err := q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         adminID,
		OrganizationID: org.ID,
		Role:           permissions.RoleAdmin,
	})
	if err != nil {
		t.Fatal(err)
	}
	memberMembership, err := q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		UserID:         memberID,
		OrganizationID: org.ID,
		Role:           permissions.RoleMember,
	})
	if err != nil {
		t.Fatal(err)
	}

	if appErr := tuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + ownerID.String(),
			Relation: permissions.RoleOwner,
			Object:   "organization:" + org.ID.String(),
		},
		{
			User:     "user:" + adminID.String(),
			Relation: permissions.RoleAdmin,
			Object:   "organization:" + org.ID.String(),
		},
		{
			User:     "user:" + memberID.String(),
			Relation: permissions.RoleMember,
			Object:   "organization:" + org.ID.String(),
		},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	t.Run("Owner can list all memberships", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		res, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 50,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(res.OrganizationMemberships) != 3 {
			t.Fatalf("expected 3 memberships but got %d", len(res.OrganizationMemberships))
		}

		foundIDs := map[string]bool{}
		for _, membership := range res.OrganizationMemberships {
			foundIDs[membership.Id] = true
		}
		if !foundIDs[ownerMembership.ID.String()] {
			t.Errorf("expected owner membership %s to be present", ownerMembership.ID.String())
		}
		if !foundIDs[adminMembership.ID.String()] {
			t.Errorf("expected admin membership %s to be present", adminMembership.ID.String())
		}
		if !foundIDs[memberMembership.ID.String()] {
			t.Errorf("expected member membership %s to be present", memberMembership.ID.String())
		}
	})

	t.Run("Member can list memberships with pagination", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, memberID.String(),
		))

		pageOne, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 2,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(pageOne.OrganizationMemberships) != 2 {
			t.Fatalf("expected 2 memberships on page 1 but got %d", len(pageOne.OrganizationMemberships))
		}

		pageTwo, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  2,
				Limit: 2,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(pageTwo.OrganizationMemberships) != 1 {
			t.Fatalf("expected 1 membership on page 2 but got %d", len(pageTwo.OrganizationMemberships))
		}
	})

	t.Cleanup(func() {
		_ = tuppleManager.Delete(context.Background(), []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + ownerID.String(),
				Relation: permissions.RoleOwner,
				Object:   "organization:" + org.ID.String(),
			},
			{
				User:     "user:" + adminID.String(),
				Relation: permissions.RoleAdmin,
				Object:   "organization:" + org.ID.String(),
			},
			{
				User:     "user:" + memberID.String(),
				Relation: permissions.RoleMember,
				Object:   "organization:" + org.ID.String(),
			},
		})
		q.DeleteOrganizationMembershipHard(context.Background(), memberMembership.ID)
		q.DeleteOrganizationMembershipHard(context.Background(), adminMembership.ID)
		q.DeleteOrganizationMembershipHard(context.Background(), ownerMembership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}

func Test_GetOrganizationMembershipsByOrgID_Integration_Failure(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := orgmembership.New(q, servicetestutils.MockUserServiceClient, fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)

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
	if appErr := tuppleManager.Write(ctx, []client.ClientTupleKey{
		{
			User:     "user:" + ownerID.String(),
			Relation: permissions.RoleOwner,
			Object:   "organization:" + org.ID.String(),
		},
	}); appErr != nil {
		t.Fatal(appErr)
	}

	t.Run("Without user metadata", func(t *testing.T) {
		_, err := service.GetOrganizationMembershipsByOrgID(context.Background(), &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 50,
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeInternal)) {
			t.Errorf("expected internal error, got %s", err.Error())
		}
	})

	t.Run("With invalid organization id", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: "invalid-id",
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 50,
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected validation error, got %s", err.Error())
		}
	})

	t.Run("With invalid pagination", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		_, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  0,
				Limit: 50,
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected validation error, got %s", err.Error())
		}
	})

	t.Run("Without organization permission", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, uuid.NewString(),
		))

		_, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 50,
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodePermissionDenied)) {
			t.Errorf("expected permission denied error, got %s", err.Error())
		}
	})

	t.Cleanup(func() {
		_ = tuppleManager.Delete(context.Background(), []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + ownerID.String(),
				Relation: permissions.RoleOwner,
				Object:   "organization:" + org.ID.String(),
			},
		})
		q.DeleteOrganizationMembershipHard(context.Background(), ownerMembership.ID)
		q.DeleteOrganizationHard(context.Background(), org.ID)
		pool.Close()
	})
}
