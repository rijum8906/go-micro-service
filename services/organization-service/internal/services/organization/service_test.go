package organization_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/testutils"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/packages/pb/user_service/mocks"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/organization"
)

var mockUserServiceClient = &mocks.MockUserServiceClient{}

func Test_CreateOrganization_Success_Integration(t *testing.T) {
	pool := testutils.MustConnectDB()
	q := db.New(pool)
	service := organization.New(q, mockUserServiceClient)

	ctx := context.Background()

	createdBy := uuid.NewString()

	createOrg := &organizationv1.CreateOrganizationRequest{
		Name:        testutils.GenerateRandomString(5),
		Description: testutils.GenerateRandomString(20),
		Slug:        "org-1",
		CreatedBy:   createdBy,
	}

	mockUserServiceClient.On("CheckExists", ctx, &userv1.CheckExistsRequest{
		Id: createdBy,
	}).Return(&userv1.CheckExistsResponse{
		Exists: true,
	}, nil)

	org, err := service.CreateOrganization(ctx, createOrg)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		id, _ := uuid.Parse(org.Id)
		q.DeleteOrganizationHard(ctx, id)
	})
}
