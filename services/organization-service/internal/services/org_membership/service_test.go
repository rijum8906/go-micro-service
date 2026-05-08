package orgmembership_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	orgmembershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/services/service_test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../../.env"); err != nil {
		if err := godotenv.Load("../../.env.test"); err != nil {
			if os.Getenv("CI") == "" {
				panic("No .env file found")
			}
		}
	}
	os.Exit(m.Run())
}

// =============================================================================
// GetMyMemberships Tests
// =============================================================================

func Test_GetMyMemberships_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
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
	suite := testutil.NewTestSuite(t)
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

// =============================================================================
// GetMyMembership Tests
// =============================================================================

func Test_GetMyMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
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
	suite := testutil.NewTestSuite(t)
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
		suite.Q.DeleteOrganizationMembershipHard(suite.Ctx, otherMembership.ID)
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

// =============================================================================
// GetOrganizationMembershipsByOrgID Tests
// =============================================================================

func Test_GetOrganizationMembershipsByOrgID_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)
	orgSuite.CreateAdmin(t, adminID)
	orgSuite.CreateMember(t, memberID)

	t.Run("owner_can_list_all_memberships", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		res, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id: orgSuite.Org.ID.String(),
			Pagination: &corev1.PaginationRequest{
				Page:  1,
				Limit: 50,
			},
		})

		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.OrganizationMemberships, 3)

		membershipIDs := make(map[string]bool)
		for _, membership := range res.OrganizationMemberships {
			membershipIDs[membership.Id] = true
		}

		assert.True(t, membershipIDs[orgSuite.OwnerMembership.ID.String()], "owner membership missing")
		assert.True(t, membershipIDs[orgSuite.AdminMembership.ID.String()], "admin membership missing")
		assert.True(t, membershipIDs[orgSuite.MemberMembership.ID.String()], "member membership missing")
	})

	t.Run("pagination_returns_correct_page_sizes", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))

		pageOne, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id:         orgSuite.Org.ID.String(),
			Pagination: &corev1.PaginationRequest{Page: 1, Limit: 2},
		})
		require.NoError(t, err)
		assert.Len(t, pageOne.OrganizationMemberships, 2, "page 1 should have 2 memberships")

		pageTwo, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id:         orgSuite.Org.ID.String(),
			Pagination: &corev1.PaginationRequest{Page: 2, Limit: 2},
		})
		require.NoError(t, err)
		assert.Len(t, pageTwo.OrganizationMemberships, 1, "page 2 should have 1 membership")

		// Verify no overlap
		pageOneIDs := make(map[string]bool)
		for _, m := range pageOne.OrganizationMemberships {
			pageOneIDs[m.Id] = true
		}
		for _, m := range pageTwo.OrganizationMemberships {
			assert.False(t, pageOneIDs[m.Id], "membership should not appear on both pages")
		}
	})

	t.Run("member_can_list_memberships", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, orgSuite.MemberMembership.UserID.String(),
		))

		res, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
			Id:         orgSuite.Org.ID.String(),
			Pagination: &corev1.PaginationRequest{Page: 1, Limit: 50},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, res.OrganizationMemberships, "member should be able to list memberships")
	})
}

func Test_GetOrganizationMembershipsByOrgID_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		orgID         string
		pagination    *corev1.PaginationRequest
		expectedError string
	}{
		{
			name: "missing_user_metadata_returns_internal_error",
			setupContext: func() context.Context {
				return context.Background()
			},
			orgID:         orgSuite.Org.ID.String(),
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeInternal),
		},
		{
			name: "invalid_organization_id_format_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         "invalid-org-id",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_pagination_page_zero_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
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
			orgID:         orgSuite.Org.ID.String(),
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
			orgID:         orgSuite.Org.ID.String(),
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
			orgID:         orgSuite.Org.ID.String(),
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 1000},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "user_without_organization_permission_returns_permission_denied",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name: "non_existent_organization_id_returns_not_found",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         uuid.New().String(),
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetOrganizationMembershipsByOrgID(ctx, &corev1.IDWithPaginationReq{
				Id:         tc.orgID,
				Pagination: tc.pagination,
			})

			require.Error(t, err, "expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedError,
				"expected error type %s, got %v", tc.expectedError, err)
		})
	}
}

// =============================================================================
// GetOrganizationMembershipsByRole Tests
// =============================================================================

func Test_GetOrganizationMembershipsByRole_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)
	orgSuite.CreateAdmin(t, adminID)
	orgSuite.CreateMember(t, memberID)

	testCases := []struct {
		name          string
		userID        string
		filterRole    string
		expectedID    string
		expectedCount int
	}{
		{
			name:          "owner_can_filter_admin_memberships",
			userID:        ownerID.String(),
			filterRole:    permissions.RoleAdmin,
			expectedID:    orgSuite.AdminMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "owner_can_filter_member_memberships",
			userID:        ownerID.String(),
			filterRole:    permissions.RoleMember,
			expectedID:    orgSuite.MemberMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "owner_can_filter_owner_memberships",
			userID:        ownerID.String(),
			filterRole:    permissions.RoleOwner,
			expectedID:    orgSuite.OwnerMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "admin_can_filter_member_memberships",
			userID:        orgSuite.AdminMembership.UserID.String(),
			filterRole:    permissions.RoleMember,
			expectedID:    orgSuite.MemberMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "member_can_filter_their_own_role",
			userID:        orgSuite.MemberMembership.UserID.String(),
			filterRole:    permissions.RoleMember,
			expectedID:    orgSuite.MemberMembership.ID.String(),
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				dto.MetaUserIDKey, tc.userID,
			))

			res, err := service.GetOrganizationMembershipsByRole(ctx, &orgmembershipv1.GetOrgMembershipsByRoleReq{
				OrganizationId: orgSuite.Org.ID.String(),
				Role:           tc.filterRole,
				Pagination: &corev1.PaginationRequest{
					Page:  1,
					Limit: 50,
				},
			})

			require.NoError(t, err)
			require.Len(t, res.OrganizationMemberships, tc.expectedCount,
				"expected %d membership(s) for role %s", tc.expectedCount, tc.filterRole)
			assert.Equal(t, tc.expectedID, res.OrganizationMemberships[0].Id,
				"membership ID mismatch for role %s", tc.filterRole)
		})
	}
}

func Test_GetOrganizationMembershipsByRole_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		orgID         string
		role          string
		pagination    *corev1.PaginationRequest
		expectedError string
	}{
		{
			name: "missing_user_metadata_returns_internal_error",
			setupContext: func() context.Context {
				return context.Background()
			},
			orgID:         orgSuite.Org.ID.String(),
			role:          permissions.RoleOwner,
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeInternal),
		},
		{
			name: "invalid_organization_id_format_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         "invalid-org-id",
			role:          permissions.RoleOwner,
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_pagination_page_zero_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			role:          permissions.RoleOwner,
			pagination:    &corev1.PaginationRequest{Page: 0, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_role_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			role:          "invalid_role",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "user_without_permission_returns_permission_denied",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			role:          permissions.RoleOwner,
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name: "non_existent_organization_returns_not_found",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         uuid.New().String(),
			role:          permissions.RoleOwner,
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetOrganizationMembershipsByRole(ctx, &orgmembershipv1.GetOrgMembershipsByRoleReq{
				OrganizationId: tc.orgID,
				Role:           tc.role,
				Pagination:     tc.pagination,
			})

			require.Error(t, err, "expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedError,
				"expected error type %s, got %v", tc.expectedError, err)
		})
	}
}

// =============================================================================
// GetOrganizationMembershipsByStatus Tests
// =============================================================================

func Test_GetOrganizationMembershipsByStatus_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()

	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Create admin and member manually to control statuses
	adminMembership, err := suite.Q.CreateOrganizationMembership(suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         adminID,
		OrganizationID: orgSuite.Org.ID,
		Role:           permissions.RoleAdmin,
	})
	require.NoError(t, err)

	memberMembership, err := suite.Q.CreateOrganizationMembership(suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         memberID,
		OrganizationID: orgSuite.Org.ID,
		Role:           permissions.RoleMember,
	})
	require.NoError(t, err)

	// Update admin status to suspended
	adminMembership, err = suite.Q.UpdateOrganizationMembershipStatus(suite.Ctx, db.UpdateOrganizationMembershipStatusParams{
		ID:     adminMembership.ID,
		Status: "suspended",
	})
	require.NoError(t, err)

	// Delete member membership
	err = suite.Q.DeleteOrganizationMembership(suite.Ctx, db.DeleteOrganizationMembershipParams{
		ID:        memberMembership.ID,
		DeletedBy: orgSuite.OwnerMembership.ID,
	})
	require.NoError(t, err)

	// Write FGA tuples
	appErr := suite.TuppleManager.Write(suite.Ctx, []client.ClientTupleKey{
		{User: "user:" + adminID.String(), Relation: permissions.RoleAdmin, Object: "organization:" + orgSuite.Org.ID.String()},
		{User: "user:" + memberID.String(), Relation: permissions.RoleMember, Object: "organization:" + orgSuite.Org.ID.String()},
	})
	require.Nil(t, appErr)

	t.Cleanup(func() {
		suite.TuppleManager.Delete(suite.Ctx, []client.ClientTupleKeyWithoutCondition{
			{User: "user:" + adminID.String(), Relation: permissions.RoleAdmin, Object: "organization:" + orgSuite.Org.ID.String()},
			{User: "user:" + memberID.String(), Relation: permissions.RoleMember, Object: "organization:" + orgSuite.Org.ID.String()},
		})
		suite.Q.DeleteOrganizationMembershipHard(suite.Ctx, memberMembership.ID)
		suite.Q.DeleteOrganizationMembershipHard(suite.Ctx, adminMembership.ID)
	})

	testCases := []struct {
		name          string
		status        string
		expectedID    string
		expectedCount int
	}{
		{
			name:          "owner_can_filter_suspended_memberships",
			status:        "suspended",
			expectedID:    adminMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "owner_can_filter_deleted_memberships",
			status:        "left",
			expectedID:    memberMembership.ID.String(),
			expectedCount: 1,
		},
		{
			name:          "owner_can_filter_active_memberships",
			status:        "active",
			expectedID:    orgSuite.OwnerMembership.ID.String(),
			expectedCount: 1, // Only owner is active
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				dto.MetaUserIDKey, ownerID.String(),
			))

			res, err := service.GetOrganizationMembershipsByStatus(ctx, &orgmembershipv1.GetOrgMembershipsByStatusReq{
				OrganizationId: orgSuite.Org.ID.String(),
				Status:         tc.status,
				Pagination: &corev1.PaginationRequest{
					Page:  1,
					Limit: 50,
				},
			})

			require.NoError(t, err)
			require.Len(t, res.OrganizationMemberships, tc.expectedCount,
				"expected %d membership(s) with status %s", tc.expectedCount, tc.status)

			if tc.expectedCount == 1 {
				assert.Equal(t, tc.expectedID, res.OrganizationMemberships[0].Id,
					"membership ID mismatch for status %s", tc.status)
			}
		})
	}
}

func Test_GetOrganizationMembershipsByStatus_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		orgID         string
		status        string
		pagination    *corev1.PaginationRequest
		expectedError string
	}{
		{
			name: "missing_user_metadata_returns_internal_error",
			setupContext: func() context.Context {
				return context.Background()
			},
			orgID:         orgSuite.Org.ID.String(),
			status:        "active",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeInternal),
		},
		{
			name: "invalid_organization_id_format_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         "invalid-org-id",
			status:        "active",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_pagination_page_zero_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			status:        "active",
			pagination:    &corev1.PaginationRequest{Page: 0, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "invalid_status_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			status:        "invalid_status",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeValidation),
		},
		{
			name: "user_without_permission_returns_permission_denied",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			orgID:         orgSuite.Org.ID.String(),
			status:        "active",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name: "non_existent_organization_returns_not_found",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			orgID:         uuid.New().String(),
			status:        "active",
			pagination:    &corev1.PaginationRequest{Page: 1, Limit: 50},
			expectedError: string(apperror.CodeNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetOrganizationMembershipsByStatus(ctx, &orgmembershipv1.GetOrgMembershipsByStatusReq{
				OrganizationId: tc.orgID,
				Status:         tc.status,
				Pagination:     tc.pagination,
			})

			require.Error(t, err, "expected error for test case: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedError,
				"expected error type %s, got %v", tc.expectedError, err)
		})
	}
}

// =============================================================================
// GetOrganizationMembership Tests
// =============================================================================

func Test_GetOrganizationMembership_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Create a deleted membership for testing
	deletedMemberID := uuid.New()
	deletedMembership, _ := suite.Q.CreateOrganizationMembership(suite.Ctx, db.CreateOrganizationMembershipParams{
		UserID:         deletedMemberID,
		OrganizationID: orgSuite.Org.ID,
		Role:           permissions.RoleMember,
	})
	suite.Q.DeleteOrganizationMembership(suite.Ctx, db.DeleteOrganizationMembershipParams{
		ID:        deletedMembership.ID,
		DeletedBy: orgSuite.OwnerMembership.ID,
	})

	testCases := []struct {
		name          string
		setupContext  func() context.Context
		requestID     string
		expectedError string
	}{
		{
			name: "missing_user_metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			requestID:     orgSuite.OwnerMembership.ID.String(),
			expectedError: string(apperror.CodeInternal),
		},
		{
			name: "non_existent_membership_id",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			requestID:     uuid.New().String(),
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name: "deleted_membership",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			requestID:     deletedMembership.ID.String(),
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name: "invalid_user_id_in_metadata",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, "invalid-uuid",
				))
			},
			requestID:     orgSuite.OwnerMembership.ID.String(),
			expectedError: string(apperror.CodePermissionDenied),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, err := service.GetOrganizationMembership(ctx, &corev1.IDRequest{
				Id: tc.requestID,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_GetOrganizationMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	// Create organization
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		dto.MetaUserIDKey, ownerID.String(),
	))

	membership, err := service.GetOrganizationMembership(ctx, &corev1.IDRequest{
		Id: orgSuite.OwnerMembership.ID.String(),
	})

	require.NoError(t, err)
	assert.Equal(t, orgSuite.OwnerMembership.ID.String(), membership.Id)
	assert.Equal(t, permissions.RoleOwner, membership.Role)
	assert.Equal(t, orgSuite.Org.ID.String(), membership.OrganizationId)
}

// =============================================================================
// SendInvitation Tests
// =============================================================================

func Test_SendInvitation_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)

	testCases := []struct {
		name           string
		setupContext   func() context.Context
		organizationID string
		email          string
		emailExists    bool
		role           string
		expectedError  string
	}{
		{
			name:           "whithout_contexted_user_info_returns_error",
			setupContext:   func() context.Context { return suite.Ctx },
			organizationID: orgSuite.Org.ID.String(),
			email:          testutils.GenerateRandomEmail(),
			emailExists:    true,
			role:           "intern",
			expectedError:  string(apperror.CodeInternal),
		},
		{
			name: "non_existing_email_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, ownerID.String(),
				))
			},
			organizationID: orgSuite.Org.ID.String(),
			email:          testutils.GenerateRandomEmail(),
			emailExists:    false,
			role:           "intern",
			expectedError:  string(apperror.CodeNotFound),
		},
		{
			name: "member_cannot_invite_member",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, memberID.String(),
				))
			},
			organizationID: orgSuite.Org.ID.String(),
			email:          testutils.GenerateRandomEmail(),
			emailExists:    true,
			role:           "member",
			expectedError:  string(apperror.CodePermissionDenied),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			servicetestutils.MockUserServiceClient.On("CheckEmailExists", ctx, &corev1.EmailRequest{Email: tc.email}).Return(&userv1.CheckExistsResponse{Exists: tc.emailExists}, nil)
			_, err := service.SendInvitation(ctx, &org_membershipv1.SendInvitationRequest{
				Email:          tc.email,
				Role:           tc.role,
				OrganizationId: tc.organizationID,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_SendInvitation_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)
	internEmail := testutils.GenerateRandomEmail()

	ctx := metadata.NewIncomingContext(suite.Ctx, metadata.Pairs(
		dto.MetaUserIDKey, ownerID.String(),
	))

	servicetestutils.MockUserServiceClient.On("CheckEmailExists", ctx, &corev1.EmailRequest{Email: internEmail}).Return(&userv1.CheckExistsResponse{Exists: true}, nil)

	_, err := service.SendInvitation(ctx, &org_membershipv1.SendInvitationRequest{
		Email:          internEmail,
		Role:           "intern",
		OrganizationId: orgSuite.Org.ID.String(),
	})
	require.NoError(t, err)
}
