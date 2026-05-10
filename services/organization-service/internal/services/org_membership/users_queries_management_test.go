package orgmembership_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func Test_LeaveOrganization_Validation(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name: "invalid_organization_id_format_returns_validation_error",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			orgMemID:      "invalid_id",
			tokenScope:    string(token.TokenScopeAdmin),
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_token_scope_returns_validation_error",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			orgMemID:      uuid.NewString(),
			tokenScope:    "invalid_scope",
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.LeaveOrganization(tc.setupContext(), &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_LeaveOrganization_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name: "user_without_user_info_ctx_returns_error",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    string(token.TokenScopeAdmin),
			expectedError: string(apperror.CodeInternal),
		},
		{
			name: "user_without_organization_permission_returns_permission_denied",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, adminID.String(),
				))
			},
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    string(token.TokenScopeAdmin),
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name: "user_with_another_org_membership_id_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgMemID:      uuid.NewString(),
			tokenScope:    string(token.TokenScopeLeaveOrganization),
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name: "user_with_other_user_id_in_ctx_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    string(token.TokenScopeLeaveOrganization),
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name: "owner_leaves_organization_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			tokenScope:    string(token.TokenScopeLeaveOrganization),
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.LeaveOrganization(tc.setupContext(), &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
