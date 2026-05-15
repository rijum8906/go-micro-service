package orgmembership_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

var (
	fgaClient    *coreopenfga.Client
	storeManager coreopenfga.StoreManager
)

func TestMain(m *testing.M) {
	f, err := coreopenfga.NewClient("http://localhost:9000")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create OpenFGA client:", err)
		os.Exit(1)
	}
	fgaClient = f

	ctx := context.Background()
	sm := coreopenfga.NewStoreManager(f.Client)
	if _, err = sm.Create(ctx, testutils.GenerateRandomString(10)); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create store:", err)
		os.Exit(1)
	}
	storeManager = sm
	f.StoreID = sm.GetStoreID()

	mm := coreopenfga.NewModelManager(f.Client, sm)
	if err = mm.Write(ctx, "organization"); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write model:", err)
		os.Exit(1)
	}
	f.AuthorizationModelID = mm.GetAuthorizationModelID()

	code := m.Run()

	// Best-effort cleanup — ignore errors so a failed store delete
	// doesn't mask a real test failure.
	_ = storeManager.Delete(ctx)
	os.Exit(code)
}

func Test_GetMyMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)
	orgSuite.CreateAdmin(t, adminID)
	orgSuite.CreateMember(t, memberID)

	testCases := []struct {
		name         string
		userID       string
		membershipID string
		expectedRole string
	}{
		{
			name:         "owner_can_view_their_membership",
			userID:       ownerID.String(),
			membershipID: orgSuite.OwnerMembership.ID.String(),
			expectedRole: permissions.RoleOwner,
		},
		{
			name:         "admin_can_view_their_membership",
			userID:       adminID.String(),
			membershipID: orgSuite.AdminMembership.ID.String(),
			expectedRole: permissions.RoleAdmin,
		},
		{
			name:         "member_can_view_their_membership",
			userID:       memberID.String(),
			membershipID: orgSuite.MemberMembership.ID.String(),
			expectedRole: permissions.RoleMember,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				dto.MetaUserIDKey, tc.userID,
			))

			res, err := service.GetMyMembership(ctx, &corev1.IDRequest{
				Id: tc.membershipID,
			})

			require.NoError(t, err)
			require.NotNil(t, res)

			assert.Equal(t, tc.membershipID, res.Id)
			assert.Equal(t, tc.userID, res.UserId)
			assert.Equal(t, orgSuite.Org.ID.String(), res.OrganizationId)
			assert.Equal(t, tc.expectedRole, res.Role)
		})
	}
}

func Test_GetMyMembership_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	otherUserID := uuid.New()

	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Create another user as member
	otherMembership, err := suite.Q.CreateOrganizationMembership(suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         otherUserID,
		OrganizationID: orgSuite.Org.ID,
		Role:           permissions.RoleMember,
	})
	require.NoError(t, err)

	// Add FGA tuple for other user
	appErr := suite.TuppleManager.Write(suite.Ctx, []client.ClientTupleKey{
		{
			User:     "user:" + otherUserID.String(),
			Relation: permissions.RoleMember,
			Object:   "organization:" + orgSuite.Org.ID.String(),
		},
	})
	require.Nil(t, appErr)

	t.Cleanup(func() {
		suite.TuppleManager.Delete(suite.Ctx, []client.ClientTupleKeyWithoutCondition{
			{
				User:     "user:" + otherUserID.String(),
				Relation: permissions.RoleMember,
				Object:   "organization:" + orgSuite.Org.ID.String(),
			},
		})
		suite.Q.HardDeleteOrganizationMembership(suite.Ctx, otherMembership.ID)
	})

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		requestID     string
		expectedError string
	}{
		{
			name: "missing_user_metadata_returns_unauthenticated",
			setupContext: func() context.Context {
				return context.Background()
			},
			requestID:     orgSuite.OwnerMembership.ID.String(),
			expectedError: string(apperror.CodeUnAuthenticated),
		},
		{
			name: "invalid_membership_id_format_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			requestID:     "not-a-valid-uuid",
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "non_existent_membership_id_returns_not_found",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			requestID:     uuid.New().String(),
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name: "accessing_another_users_membership_returns_permission_denied",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			requestID:     otherMembership.ID.String(),
			expectedError: string(apperror.CodePermissionDenied),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetMyMembership(ctx, &corev1.IDRequest{Id: tc.requestID})

			require.Error(t, err, "expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedError,
				"expected error type %s, got %v", tc.expectedError, err)
		})
	}
}

func Test_GetMyMemberships_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)

	t.Run("owner_can_view_their_membership", func(t *testing.T) {
		orgSuite.CreateOwner(t, ownerID)
		expectedMembershipID := orgSuite.OwnerMembership.ID.String()

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		memberships, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})

		require.NoError(t, err)
		require.Len(t, memberships.OrganizationMemberships, 1)

		actual := memberships.OrganizationMemberships[0]
		assert.Equal(t, expectedMembershipID, actual.Id)
		assert.Equal(t, permissions.RoleOwner, actual.Role)
		assert.Equal(t, orgSuite.Org.ID.String(), actual.OrganizationId)
	})

	t.Run("admin_can_view_their_membership", func(t *testing.T) {
		adminID := uuid.New()
		orgSuite.CreateAdmin(t, adminID)
		expectedMembershipID := orgSuite.AdminMembership.ID.String()

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, adminID.String(),
		))

		memberships, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})

		require.NoError(t, err)
		require.Len(t, memberships.OrganizationMemberships, 1)

		actual := memberships.OrganizationMemberships[0]
		assert.Equal(t, expectedMembershipID, actual.Id)
		assert.Equal(t, permissions.RoleAdmin, actual.Role)
		assert.Equal(t, orgSuite.Org.ID.String(), actual.OrganizationId)
	})

	t.Run("member_can_view_their_membership", func(t *testing.T) {
		memberID := uuid.New()
		orgSuite.CreateMember(t, memberID)
		expectedMembershipID := orgSuite.MemberMembership.ID.String()

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, memberID.String(),
		))

		memberships, err := service.GetMyMemberships(ctx, &corev1.PaginationRequest{
			Page:  1,
			Limit: 50,
		})

		require.NoError(t, err)
		require.Len(t, memberships.OrganizationMemberships, 1)

		actual := memberships.OrganizationMemberships[0]
		assert.Equal(t, expectedMembershipID, actual.Id)
		assert.Equal(t, permissions.RoleMember, actual.Role)
		assert.Equal(t, orgSuite.Org.ID.String(), actual.OrganizationId)
	})
}

func Test_GetMyMemberships_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		pagination    *corev1.PaginationRequest
		expectedError string
	}{
		{
			name: "missing_user_metadata_returns_unauthenticated",
			setupContext: func() context.Context {
				return context.Background()
			},
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeUnAuthenticated),
		},
		{
			name: "invalid_pagination_page_zero_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			pagination:    &corev1.PaginationRequest{Page: 0, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_pagination_negative_page_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			pagination:    &corev1.PaginationRequest{Page: -1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_pagination_limit_zero_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 0},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "pagination_limit_exceeds_maximum_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 1000},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_user_id_format_returns_internal_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, "not-a-valid-uuid",
				))
			},
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeInternal),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetMyMemberships(ctx, tc.pagination)

			require.Error(t, err, "expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedError,
				"expected error type %s, got %v", tc.expectedError, err)
		})
	}
}
