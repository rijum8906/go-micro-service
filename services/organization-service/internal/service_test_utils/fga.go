// Package servicetestutils
package servicetestutils

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/organization_service/models/v1"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	userv1_mock "github.com/rijum8906/relay/packages/pb/user_service/mocks"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
)

var MockUserServiceClient = &userv1_mock.MockUserServiceClient{}

func MustCreateService() (db.Querier, *pgxpool.Pool, *coreopenfga.Client) {
	apiURL := GetEnv("OPENFGA_TEST_API_URL", "FGA_TEST_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:9000"
	}
	storeID := GetEnv("OPENFGA_TEST_STORE_ID", "FGA_TEST_STORE_ID")
	authModelID := GetEnv("OPENFGA_TEST_AUTH_MODEL_ID", "FGA_TEST_AUTH_MODEL_ID")

	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("organization-service")))
	q := db.New(pool)
	fgaClient, err := coreopenfga.NewClient(apiURL)
	if err != nil {
		panic(apiURL)
	}
	fgaClient.StoreID = storeID
	fgaClient.AuthorizationModelID = authModelID

	return q, pool, fgaClient
}

func GetEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			return value
		}
	}

	return ""
}

func NormalizeCreateRequest(req *organizationv1.CreateOrganizationRequest) *organizationv1.CreateOrganizationRequest {
	// Handle nil request
	if req == nil {
		return &organizationv1.CreateOrganizationRequest{
			Name:        testutils.GenerateRandomString(5),
			Description: testutils.GenerateRandomString(20),
			Slug:        strings.ToLower(testutils.GenerateRandomString(5)),
			CreatedBy:   uuid.New().String(),
		}
	}

	// Create a copy to avoid modifying original
	normalized := &organizationv1.CreateOrganizationRequest{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		CreatedBy:   req.CreatedBy,
	}

	// Apply defaults for empty fields
	if normalized.Name == "" {
		normalized.Name = testutils.GenerateRandomString(5)
	}
	if normalized.Description == "" {
		normalized.Description = testutils.GenerateRandomString(20)
	}
	if normalized.Slug == "" {
		normalized.Slug = strings.ToLower(testutils.GenerateRandomString(5))
	}
	if normalized.CreatedBy == "" {
		normalized.CreatedBy = uuid.New().String()
	}

	return normalized
}

func MatchOrganization(a, b *modelsv1.Organization) bool {
	return a.Id == b.Id &&
		a.Name == b.Name &&
		a.Slug == b.Slug &&
		a.Status == b.Status
}

func MustCreateOrg(ctx context.Context, q db.Querier, req *organizationv1.CreateOrganizationRequest) *db.Organization {
	req = NormalizeCreateRequest(req)

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		panic(err)
	}
	MockUserServiceClient.On("CheckExists", ctx, &corev1.IDRequest{
		Id: req.CreatedBy,
	}).Return(&userv1.CheckExistsResponse{
		Exists: true,
	}, nil)

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
		Slug:        req.Slug,
		CreatedBy:   createdBy,
	})
	if err != nil {
		panic(err)
	}

	return &org
}
