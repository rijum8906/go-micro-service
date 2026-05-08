package organization_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/service_test_utils"
	"github.com/rijum8906/relay/services/organization-service/internal/services/organization"
	grpcmetadata "google.golang.org/grpc/metadata"
)

// NOTE: do not perform any validation test here, as it has already been tested in validation_test.go

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

func Test_CreateOrganization_Success_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)

	ctx := context.Background()

	createdBy := uuid.New()

	createOrg := &organizationv1.CreateOrganizationRequest{
		Name:        testutils.GenerateRandomString(5),
		Description: testutils.GenerateRandomString(20),
		Slug:        strings.ToLower(testutils.GenerateRandomString(5)),
		CreatedBy:   createdBy.String(),
	}

	servicetestutils.MockUserServiceClient.On("CheckExists", ctx, &userv1.CheckExistsRequest{
		Id: createdBy.String(),
	}).Return(&userv1.CheckExistsResponse{
		Exists: true,
	}, nil)

	org, err := service.CreateOrganization(ctx, createOrg)
	if err != nil {
		t.Fatal(err)
	}

	orgDups, err := q.GetOrganizationMembershipsByUserID(ctx, db.GetOrganizationMembershipsByUserIDParams{
		UserID: createdBy,
		Limit:  3,
		Offset: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Only one memebr should be there that is the owner
	if len(orgDups) != 1 {
		t.Errorf("expected only 1 memebr (owner) but got %d", len(orgDups))
	}

	if orgDups[0].OrganizationID.String() != org.Id {
		t.Errorf("expected organization id to be %s but got %s", org.Id, orgDups[0].OrganizationID.String())
	}
	if orgDups[0].Role != "owner" {
		t.Errorf("expected the role to be owner but got %s", orgDups[0].Status)
	}

	// Check the relation in fga client
	check, appErr := tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + createdBy.String(),
		Relation: permissions.RoleOwner,
		Object:   "organization:" + org.Id,
	})
	if appErr != nil {
		t.Fatal(appErr)
	}
	if !*check.Allowed {
		t.Errorf("expected the relation to be allowed but got %t", *check.Allowed)
	}

	t.Cleanup(func() {
		orgID, _ := uuid.Parse(org.Id)
		q.DeleteOrganizationHard(ctx, orgID)
		// On organization delete it's members will be automatically deleted
		pool.Close()
	})
}

func Test_GetOrganization_Success_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	org := mustCreateOrg(ctx, service, &organizationv1.CreateOrganizationRequest{
		Slug: "org-4",
	})

	fetchedOrg, err := service.GetOrganization(ctx, &corev1.IDRequest{
		Id: org.Id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !servicetestutils.MatchOrganization(fetchedOrg, org) {
		t.Errorf("expected organization to be %v but got %v", org, fetchedOrg)
	}

	fetchedOrg, err = service.GetOrganizationBySlug(ctx, &organizationv1.GetOrganizationBySlugRequest{
		Slug: org.Slug,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !servicetestutils.MatchOrganization(fetchedOrg, org) {
		t.Errorf("expected organization to be %v but got %v", org, fetchedOrg)
	}

	ctx = grpcmetadata.NewIncomingContext(ctx, grpcmetadata.Pairs(
		dto.MetaUserIDKey, org.CreatedBy,
	))
	fetchedOrgs, err := service.GetOrganizationsListByCreatedBy(ctx, &corev1.EmptyRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if len(fetchedOrgs.Organizations) != 1 {
		t.Errorf("expected 1 organization but got %d", len(fetchedOrgs.Organizations))
	}
	if fetchedOrgs.Organizations[0].Id != org.Id {
		t.Errorf("expected organization id to be %s but got %s", org.Id, fetchedOrgs.Organizations[0].Id)
	}

	t.Cleanup(func() {
		id, _ := uuid.Parse(org.Id)
		q.DeleteOrganizationHard(ctx, id)
		pool.Close()
	})
}

func Test_GetOrganization_Failure_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	org := mustCreateOrg(ctx, service, &organizationv1.CreateOrganizationRequest{
		Slug: "org-5",
	})
	id, _ := uuid.Parse(org.Id)
	err := q.DeleteOrganizationHard(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.GetOrganization(ctx, id)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if strings.Contains(err.Error(), string(apperror.CodeNotFound)) {
			t.Error("expected error code not found, got something else")
		}
	}

	t.Cleanup(func() {
		pool.Close()
	})
}

func Test_ChangeOrganizationOwnership_Success_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	createdBy := uuid.New()
	newOwner := uuid.New()

	// Update Context With UserInfo
	ctx = grpcmetadata.NewIncomingContext(ctx, grpcmetadata.Pairs(
		dto.MetaUserIDKey, createdBy.String(),
	))

	servicetestutils.MockUserServiceClient.On("CheckExists", ctx, &userv1.CheckExistsRequest{
		Id: newOwner.String(),
	}).Return(&userv1.CheckExistsResponse{
		Exists: true,
	}, nil)

	org := mustCreateOrg(ctx, service, &organizationv1.CreateOrganizationRequest{
		Name:        testutils.GenerateRandomString(5),
		Description: testutils.GenerateRandomString(20),
		Slug:        strings.ToLower(testutils.GenerateRandomString(5)),
		CreatedBy:   createdBy.String(),
	})

	orgID, err := uuid.Parse(org.Id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.ChangeOrganizationOwnership(ctx, &organizationv1.ChangeOrganizationOwnershipRequest{
		OrganizationId: org.Id,
		NewOwnerId:     newOwner.String(),
		TokenScope:     string(token.TokenScopeChangeOrganizationOwner),
	})
	if err != nil {
		t.Fatal(err)
	}

	fetchedOrg, err := q.GetOrganization(ctx, orgID)
	if err != nil {
		t.Fatal(err)
	}
	if fetchedOrg.CreatedBy.String() != newOwner.String() {
		t.Errorf("expected organization owner to be %s but got %s", newOwner.String(), fetchedOrg.CreatedBy.String())
	}

	// Try to again change the ownership

	ctx = grpcmetadata.NewIncomingContext(ctx, grpcmetadata.Pairs(
		dto.MetaUserIDKey, createdBy.String(),
	))
	servicetestutils.MockUserServiceClient.On("CheckExists", ctx, &userv1.CheckExistsRequest{
		Id: newOwner.String(),
	}).Return(&userv1.CheckExistsResponse{
		Exists: true,
	}, nil)

	_, err = service.ChangeOrganizationOwnership(ctx, &organizationv1.ChangeOrganizationOwnershipRequest{
		OrganizationId: org.Id,
		NewOwnerId:     newOwner.String(),
		TokenScope:     string(token.TokenScopeChangeOrganizationOwner),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	t.Cleanup(func() {
		q.DeleteOrganizationHard(context.Background(), orgID)
		pool.Close()
	})
}

func Test_DeleteOrganization_Success_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)

	ctx := context.Background()

	// Create organization
	org := mustCreateOrg(ctx, service, &organizationv1.CreateOrganizationRequest{})
	id, err := uuid.Parse(org.Id)
	if err != nil {
		t.Fatal(err)
	}

	// Update Context With UserInfo
	ctx = grpcmetadata.NewIncomingContext(ctx, grpcmetadata.Pairs(
		dto.MetaUserIDKey, org.CreatedBy,
	))

	// Delete organization
	res, err := service.DeleteOrganization(ctx, &corev1.IDAndScopedTokenRequest{
		Id:         org.Id,
		TokenScope: string(token.TokenScopeDeleteOrganization),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Errorf("expected success to be true but got false")
	}

	// Fetch Deleted org
	deletedOrg, err := q.GetDeletedOrganization(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if deletedOrg.DeletedBy.String() != org.CreatedBy {
		t.Errorf("expected deleted by to be %s but got %s", org.CreatedBy, deletedOrg.DeletedBy.String())
	}
	if deletedOrg.DeletedAt.Time.After(time.Now()) {
		t.Errorf("deleted_at should be in the past (already deleted), but got future time: %v", deletedOrg.DeletedAt.Time)
	}

	check, appErr := tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + org.CreatedBy,
		Relation: permissions.RoleOwner,
		Object:   "organization:" + org.Id,
	})
	if appErr != nil {
		t.Fatal(appErr)
	}
	if *check.Allowed {
		t.Errorf("expected deleted organization owner tuple to be removed")
	}

	// Cleanup
	t.Cleanup(func() {
		q.DeleteOrganizationHard(context.Background(), id)
		pool.Close()
	})
}

func Test_DeleteOrganization_Failure_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	// Create Organization
	org := mustCreateOrg(ctx, service, nil)

	t.Run("Without proper context", func(t *testing.T) {
		_, err := service.DeleteOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         org.Id,
			TokenScope: string(token.TokenScopeDeleteOrganization),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeInternal)) {
			t.Errorf("expected error code validation, got something else")
		}
	})

	t.Run("With invalid organization id", func(t *testing.T) {
		ctx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs(
			dto.MetaUserIDKey, org.CreatedBy,
		))

		_, err := service.DeleteOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         uuid.NewString(),
			TokenScope: string(token.TokenScopeDeleteOrganization),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeNotFound)) {
			t.Errorf("expected error code not found, got something error %s", err.Error())
		}
	})

	t.Run("Delete with wrong token scope", func(t *testing.T) {
		ctx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs(
			dto.MetaUserIDKey, org.CreatedBy,
		))

		_, err := service.DeleteOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         org.Id,
			TokenScope: string(token.TokenScopeChangeOrganizationOwner),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected error code validation, got something else")
		}
	})

	t.Cleanup(func() {
		pool.Close()
	})
}

func Test_ArchiveOrganization_Success_integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)

	ctx := context.Background()

	// Create Organization
	org := mustCreateOrg(ctx, service, nil)

	// Update Context With UserInfo
	ctx = grpcmetadata.NewIncomingContext(ctx, grpcmetadata.Pairs(
		dto.MetaUserIDKey, org.CreatedBy,
	))

	// Archive
	res, err := service.ArchiveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
		Id:         org.Id,
		TokenScope: string(token.TokenScopeArchiveOrganization),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Errorf("expected success to be true but got false")
	}

	// Fetch
	id, _ := uuid.Parse(org.Id)
	archivedOrg, err := q.GetOrganization(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if archivedOrg.Status != "archived" {
		t.Errorf("expected archived_at to be not nil but got nil")
	}
	if archivedOrg.ArchivedAt.Time.After(time.Now()) {
		t.Errorf("archived_at should be in the past (already archived), but got future time: %v", archivedOrg.ArchivedAt.Time)
	}

	check, appErr := tuppleManager.Check(ctx, client.ClientCheckRequest{
		User:     "user:" + org.CreatedBy,
		Relation: permissions.RoleOwner,
		Object:   "organization:" + org.Id,
	})
	if appErr != nil {
		t.Fatal(appErr)
	}
	if !*check.Allowed {
		t.Errorf("expected archived organization owner tuple to remain")
	}

	// Cleanup
	t.Cleanup(func() {
		q.DeleteOrganizationHard(context.Background(), id)
		pool.Close()
	})
}

func Test_ArchiveOrganization_Failure_Integration(t *testing.T) {
	q, pool, fgaClient := servicetestutils.MustCreateService()
	service := organization.New(q, servicetestutils.MockUserServiceClient, fgaClient)

	ctx := context.Background()

	org := mustCreateOrg(ctx, service, nil)

	t.Run("Without proper context", func(t *testing.T) {
		_, err := service.ArchiveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         org.Id,
			TokenScope: string(token.TokenScopeArchiveOrganization),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeInternal)) {
			t.Errorf("expected error code internal, got something else")
		}
	})

	t.Run("With invalid organization id", func(t *testing.T) {
		ctx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs(
			dto.MetaUserIDKey, org.CreatedBy,
		))

		_, err := service.ArchiveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         uuid.NewString(),
			TokenScope: string(token.TokenScopeArchiveOrganization),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeNotFound)) {
			t.Errorf("expected error code not found, got something error %s", err.Error())
		}
	})

	t.Run("Delete with wrong token scope", func(t *testing.T) {
		ctx := grpcmetadata.NewIncomingContext(context.Background(), grpcmetadata.Pairs(
			dto.MetaUserIDKey, org.CreatedBy,
		))

		_, err := service.ArchiveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			Id:         org.Id,
			TokenScope: string(token.TokenScopeChangeOrganizationOwner),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), string(apperror.CodeValidation)) {
			t.Errorf("expected error code validation, got something else")
		}
	})

	t.Cleanup(func() {
		pool.Close()
	})
}

func mustCreateOrg(ctx context.Context, service organizationv1.OrganizationServiceServer, req *organizationv1.CreateOrganizationRequest) *modelsv1.Organization {
	// Normalize request with defaults
	normalizedReq := servicetestutils.NormalizeCreateRequest(req)

	// Setup user existence mock
	servicetestutils.MockUserServiceClient.On("CheckExists", ctx, &userv1.CheckExistsRequest{
		Id: normalizedReq.CreatedBy,
	}).Return(&userv1.CheckExistsResponse{Exists: true}, nil)

	// Create organization
	org, err := service.CreateOrganization(ctx, normalizedReq)
	if err != nil {
		panic(err)
	}

	return org
}
