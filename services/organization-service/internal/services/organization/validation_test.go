package organization

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
)

// NOTE: do not test slug too much validateSlug() is already being tested
// Do not repeat already done validation test in validation_test.go

func Test_validateCreateOrganizationRequest(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req     *organizationv1.CreateOrganizationRequest
		wantErr bool
	}{
		{
			name: "invalid slug with a whitespace",
			req: &organizationv1.CreateOrganizationRequest{
				Name:        testutils.GenerateRandomString(5),
				Description: testutils.GenerateRandomString(20),
				Slug:        "org v",
				CreatedBy:   uuid.NewString(),
			},
			wantErr: true,
		},
		{
			name: "invalid created by",
			req: &organizationv1.CreateOrganizationRequest{
				Name:        testutils.GenerateRandomString(5),
				Description: testutils.GenerateRandomString(20),
				Slug:        "org-1",
				CreatedBy:   "122",
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &organizationv1.CreateOrganizationRequest{
				Name:        testutils.GenerateRandomString(5),
				Description: testutils.GenerateRandomString(20),
				Slug:        "org-1",
				CreatedBy:   uuid.NewString(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateCreateOrganizationRequest(tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateCreateOrganizationRequest() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateCreateOrganizationRequest() succeeded unexpectedly")
			}
		})
	}
}

func Test_validateChangeOwnershipRequst(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req     *organizationv1.ChangeOrganizationOwnershipRequest
		wantErr bool
	}{
		{
			name: "valida request",
			req: &organizationv1.ChangeOrganizationOwnershipRequest{
				OrganizationId: uuid.NewString(),
				NewOwnerId:     uuid.NewString(),
				TokenScope:     string(token.TokenScopeChangeOrganizationOwner),
			},
			wantErr: false,
		},
		{
			name: "invalid token scope",
			req: &organizationv1.ChangeOrganizationOwnershipRequest{
				OrganizationId: uuid.NewString(),
				NewOwnerId:     uuid.NewString(),
				TokenScope:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid organization id",
			req: &organizationv1.ChangeOrganizationOwnershipRequest{
				OrganizationId: "invalid",
				NewOwnerId:     uuid.NewString(),
				TokenScope:     string(token.TokenScopeChangeOrganizationOwner),
			},
			wantErr: true,
		},
		{
			name: "invalid new owner id",
			req: &organizationv1.ChangeOrganizationOwnershipRequest{
				OrganizationId: uuid.NewString(),
				NewOwnerId:     "invalid",
				TokenScope:     string(token.TokenScopeChangeOrganizationOwner),
			},
			wantErr: true,
		},
		{
			name: "blank ids and token scope",
			req: &organizationv1.ChangeOrganizationOwnershipRequest{
				OrganizationId: "",
				NewOwnerId:     "",
				TokenScope:     "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateChangeOwnershipRequst(tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateChangeOwnershipRequst() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateChangeOwnershipRequst() succeeded unexpectedly")
			}
		})
	}
}

func Test_validateUpdateOrganizationName(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req     *organizationv1.UpdateOrganizationNameRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &organizationv1.UpdateOrganizationNameRequest{
				OrganizationId: uuid.NewString(),
				TokenScope:     string(token.TokenScopeUpdateOrganizationName),
				Name:           testutils.GenerateRandomString(5),
				Description:    testutils.GenerateRandomString(20),
			},
			wantErr: false,
		},
		{
			name: "invalid token scope",
			req: &organizationv1.UpdateOrganizationNameRequest{
				OrganizationId: uuid.NewString(),
				TokenScope:     "invalid",
				Name:           testutils.GenerateRandomString(5),
				Description:    testutils.GenerateRandomString(20),
			},
			wantErr: true,
		},
		{
			name: "invalid organization id",
			req: &organizationv1.UpdateOrganizationNameRequest{
				OrganizationId: "invalid",
				TokenScope:     string(token.TokenScopeUpdateOrganizationName),
				Name:           testutils.GenerateRandomString(5),
				Description:    testutils.GenerateRandomString(20),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateUpdateOrganizationName(tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateUpdateOrganizationName() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateUpdateOrganizationName() succeeded unexpectedly")
			}
		})
	}
}
