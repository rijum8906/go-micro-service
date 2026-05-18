package orgmembership_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// =============================================================================
// GetOrganizationMembershipsByOrgID Tests
// =============================================================================

func Test_GetOrganizationMembership_Validation(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	tests := []struct {
		name string
		req  *corev1.IDRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "blank id",
			req: &corev1.IDRequest{
				Id: "",
			},
		},
		{
			name: "malformed uuid",
			req: &corev1.IDRequest{
				Id: "malformed-uuid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetOrganizationMembership(suite.Ctx, tt.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), string(apperror.CodeValidation))
		})
	}
}

func Test_GetOrganizationMembershipsByOrgID_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
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

func Test_GetOrganizationMembershipsByOrgID_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
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

// =============================================================================
// GetOrganizationMembershipsByRole Tests
// =============================================================================

func Test_GetOrganizationMembershipsByRole_Integration_Success(t *testing.T) {
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

			res, err := service.GetOrganizationMembershipsByRole(ctx, &org_membershipv1.GetOrgMembershipsByRoleReq{
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
	suite := testutil.NewTestSuite(t, fgaClient)
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
			_, err := service.GetOrganizationMembershipsByRole(ctx, &org_membershipv1.GetOrgMembershipsByRoleReq{
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
	suite := testutil.NewTestSuite(t, fgaClient)
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
	err = suite.Q.SoftDeleteOrganizationMembership(suite.Ctx, memberMembership.ID)
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
		suite.Q.HardDeleteOrganizationMembership(suite.Ctx, memberMembership.ID)
		suite.Q.HardDeleteOrganizationMembership(suite.Ctx, adminMembership.ID)
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

			res, err := service.GetOrganizationMembershipsByStatus(ctx, &org_membershipv1.GetOrgMembershipsByStatusReq{
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
	suite := testutil.NewTestSuite(t, fgaClient)
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
			_, err := service.GetOrganizationMembershipsByStatus(ctx, &org_membershipv1.GetOrgMembershipsByStatusReq{
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
	suite := testutil.NewTestSuite(t, fgaClient)
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
	suite.Q.SoftDeleteOrganizationMembership(suite.Ctx, deletedMembership.ID)

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
	suite := testutil.NewTestSuite(t, fgaClient)
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
