package organization_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/repositories/organization"
)

func Test_service_CreateOrganization(t *testing.T) {
	ctx := context.Background()

	pool := testutils.MustConnectDB()
	q := db.New(pool)
	repo := organization.New(q)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		orgName     string
		description string
		createdBy   uuid.UUID
		wantErr     bool
		errType     apperror.AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, appErr := repo.CreateOrganization(ctx, tt.orgName, tt.description, uuid.New())
			if appErr != nil {
				t.Errorf("CreateOrganization() error = %v, wantErr %v", appErr, tt.want2)
				return
			}
		})
	}
}
