package orgmembership_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/app/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func Test_LeaveOrganization_Validation(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
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
			tokenScope:    constants.TokenScopeLeaveOrganization,
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
	suite := testutil.NewTestSuite(t, fgaClient)
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
			tokenScope:    string(constants.TokenScopeLeaveOrganization),
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
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
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
			tokenScope:    constants.TokenScopeLeaveOrganization,
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
			tokenScope:    constants.TokenScopeLeaveOrganization,
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
			tokenScope:    constants.TokenScopeLeaveOrganization,
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

func Test_LeaveOrganization_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	t.Run("admin_success", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, adminID.String(),
		))

		_, err := service.LeaveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeLeaveOrganization,
			Id:         orgSuite.AdminMembership.ID.String(),
		})
		require.NoError(t, err)

		membership, err := suite.Q.GetOrganizationMembershipWithAllStatuses(ctx, orgSuite.AdminMembership.ID)
		require.False(t, errors.Is(err, sql.ErrNoRows))
		require.Equal(t, constants.OrgMemStatusLeft, membership.Status)
	})

	t.Run("org_with_two_owners_one_can_leave", func(t *testing.T) {
		newOwnerID := uuid.New()
		orgSuite.CreateOwner(t, newOwnerID)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			dto.MetaUserIDKey, newOwnerID.String(),
		))

		_, err := service.LeaveOrganization(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeLeaveOrganization,
			Id:         orgSuite.OwnerMembership.ID.String(),
		})
		require.NoError(t, err)
		membership, err := suite.Q.GetOrganizationMembershipWithAllStatuses(ctx, orgSuite.OwnerMembership.ID)
		require.False(t, errors.Is(err, sql.ErrNoRows))
		require.Equal(t, constants.OrgMemStatusLeft, membership.Status)
	})
}

func Test_BanOrganizationMembership_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	suspendedMemberID := uuid.New()
	orgSuite.CreateMember(t, suspendedMemberID)
	suspendedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, suspendedMembership.ID, constants.OrgMemStatusSuspended)

	removedMemberID := uuid.New()
	orgSuite.CreateMember(t, removedMemberID)
	removedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, removedMembership.ID, constants.OrgMemStatusRemoved)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "unknown_membership_returns_not_found",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      uuid.NewString(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name:          "admin_cannot_ban_self",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_ban_owner",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "member_cannot_ban_admin",
			ctx:           orgMembershipUserCtx(memberID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_ban_suspended_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      suspendedMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_ban_removed_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      removedMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.BanOrganizationMembership(tc.ctx, &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_BanOrganizationMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(adminID)

	t.Run("admin_bans_member", func(t *testing.T) {
		_, err := service.BanOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeLeaveOrganization,
			Id:         memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusBanned)
	})

	t.Run("already_banned_membership_is_idempotent", func(t *testing.T) {
		_, err := service.BanOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeLeaveOrganization,
			Id:         memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusBanned)
	})
}

func Test_UnbanOrganizationMembership_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	suspendedMemberID := uuid.New()
	orgSuite.CreateMember(t, suspendedMemberID)
	suspendedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, suspendedMembership.ID, constants.OrgMemStatusSuspended)

	leftMemberID := uuid.New()
	orgSuite.CreateMember(t, leftMemberID)
	leftMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, leftMembership.ID, constants.OrgMemStatusLeft)

	removedMemberID := uuid.New()
	orgSuite.CreateMember(t, removedMemberID)
	removedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, removedMembership.ID, constants.OrgMemStatusRemoved)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			tokenScope:    constants.TokenScopeLeaveOrganization,
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "cannot_unban_owner",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_unban_suspended_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      suspendedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_unban_left_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      leftMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_unban_removed_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      removedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.UnbanOrganizationMembership(tc.ctx, &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_UnbanOrganizationMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	bannedMemberID := uuid.New()
	orgSuite.CreateMember(t, bannedMemberID)
	bannedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, bannedMembership.ID, constants.OrgMemStatusBanned)

	activeMemberID := uuid.New()
	orgSuite.CreateMember(t, activeMemberID)
	activeMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(adminID)

	t.Run("admin_unbans_banned_member", func(t *testing.T) {
		_, err := service.UnbanOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         bannedMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, bannedMembership.ID, constants.OrgMemStatusActive)
	})

	t.Run("already_active_membership_is_idempotent", func(t *testing.T) {
		_, err := service.UnbanOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         activeMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, activeMembership.ID, constants.OrgMemStatusActive)
	})
}

func Test_SuspendOrganizationMembership_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	peerMemberID := uuid.New()
	orgSuite.CreateMember(t, peerMemberID)
	peerMembership := orgSuite.MemberMembership

	bannedMemberID := uuid.New()
	orgSuite.CreateMember(t, bannedMemberID)
	bannedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, bannedMembership.ID, constants.OrgMemStatusBanned)

	leftMemberID := uuid.New()
	orgSuite.CreateMember(t, leftMemberID)
	leftMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, leftMembership.ID, constants.OrgMemStatusLeft)

	removedMemberID := uuid.New()
	orgSuite.CreateMember(t, removedMemberID)
	removedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, removedMembership.ID, constants.OrgMemStatusRemoved)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "admin_cannot_suspend_self",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_suspend_owner",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "member_cannot_suspend_peer_member",
			ctx:           orgMembershipUserCtx(memberID),
			orgMemID:      peerMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_suspend_banned_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      bannedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_suspend_left_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      leftMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_suspend_removed_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      removedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.SuspendOrganizationMembership(tc.ctx, &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_SuspendOrganizationMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(adminID)

	t.Run("admin_suspends_member", func(t *testing.T) {
		_, err := service.SuspendOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusSuspended)
	})

	t.Run("already_suspended_membership_is_idempotent", func(t *testing.T) {
		_, err := service.SuspendOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusSuspended)
	})
}

func Test_ActivateOrganizationMembership_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	bannedMemberID := uuid.New()
	orgSuite.CreateMember(t, bannedMemberID)
	bannedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, bannedMembership.ID, constants.OrgMemStatusBanned)

	leftMemberID := uuid.New()
	orgSuite.CreateMember(t, leftMemberID)
	leftMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, leftMembership.ID, constants.OrgMemStatusLeft)

	removedMemberID := uuid.New()
	orgSuite.CreateMember(t, removedMemberID)
	removedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, removedMembership.ID, constants.OrgMemStatusRemoved)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		tokenScope    string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "admin_cannot_activate_peer_admin",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_activate_banned_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      bannedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_activate_left_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      leftMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_activate_removed_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      removedMembership.ID.String(),
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ActivateOrganizationMembership(tc.ctx, &corev1.IDAndScopedTokenRequest{
				TokenScope: tc.tokenScope,
				Id:         tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_ActivateOrganizationMembership_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	suspendedMemberID := uuid.New()
	orgSuite.CreateMember(t, suspendedMemberID)
	suspendedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, suspendedMembership.ID, constants.OrgMemStatusSuspended)

	activeMemberID := uuid.New()
	orgSuite.CreateMember(t, activeMemberID)
	activeMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(adminID)

	t.Run("admin_activates_suspended_member", func(t *testing.T) {
		_, err := service.ActivateOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         suspendedMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, suspendedMembership.ID, constants.OrgMemStatusActive)
	})

	t.Run("already_active_membership_is_idempotent", func(t *testing.T) {
		_, err := service.ActivateOrganizationMembership(ctx, &corev1.IDAndScopedTokenRequest{
			TokenScope: constants.TokenScopeUpdateOrganizationMembership,
			Id:         activeMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, activeMembership.ID, constants.OrgMemStatusActive)
	})
}

func Test_ChangeOrganizationMembershipRole_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	peerMemberID := uuid.New()
	orgSuite.CreateMember(t, peerMemberID)
	peerMembership := orgSuite.MemberMembership

	suspendedMemberID := uuid.New()
	orgSuite.CreateMember(t, suspendedMemberID)
	suspendedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, suspendedMembership.ID, constants.OrgMemStatusSuspended)

	leftMemberID := uuid.New()
	orgSuite.CreateMember(t, leftMemberID)
	leftMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, leftMembership.ID, constants.OrgMemStatusLeft)

	removedMemberID := uuid.New()
	orgSuite.CreateMember(t, removedMemberID)
	removedMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, removedMembership.ID, constants.OrgMemStatusRemoved)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		newRole       string
		tokenScope    string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "invalid_token_scope_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      memberMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    "invalid_scope",
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "wrong_token_scope_returns_permission_denied",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      memberMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeTest,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "invalid_new_role_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      memberMembership.ID.String(),
			newRole:       "invalid_role",
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "owner_role_cannot_be_assigned",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      memberMembership.ID.String(),
			newRole:       constants.OrgRoleOwner,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "admin_cannot_change_self_role",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			newRole:       constants.OrgRoleMember,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "admin_cannot_change_owner_role",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			newRole:       constants.OrgRoleMember,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "member_cannot_change_peer_member_role",
			ctx:           orgMembershipUserCtx(memberID),
			orgMemID:      peerMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_change_suspended_membership_role",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      suspendedMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_change_left_membership_role",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      leftMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "cannot_change_removed_membership_role",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      removedMembership.ID.String(),
			newRole:       constants.OrgRoleAdmin,
			tokenScope:    constants.TokenScopeUpdateOrganizationMembership,
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ChangeOrganizationMembershipRole(tc.ctx, &org_membershipv1.ChangeOrgMembershipRoleReq{
				OrganizationMembershipId: tc.orgMemID,
				NewRole:                  tc.newRole,
				TokenScope:               tc.tokenScope,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_ChangeOrganizationMembershipRole_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	alreadyMemberID := uuid.New()
	orgSuite.CreateMember(t, alreadyMemberID)
	alreadyMemberMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(ownerID)

	t.Run("owner_changes_member_to_admin", func(t *testing.T) {
		_, err := service.ChangeOrganizationMembershipRole(ctx, &org_membershipv1.ChangeOrgMembershipRoleReq{
			OrganizationMembershipId: memberMembership.ID.String(),
			NewRole:                  constants.OrgRoleAdmin,
			TokenScope:               constants.TokenScopeUpdateOrganizationMembership,
		})
		require.NoError(t, err)
		requireOrgMembershipRole(t, suite, memberMembership.ID, constants.OrgRoleAdmin)
	})

	t.Run("same_role_update_is_idempotent", func(t *testing.T) {
		_, err := service.ChangeOrganizationMembershipRole(ctx, &org_membershipv1.ChangeOrgMembershipRoleReq{
			OrganizationMembershipId: alreadyMemberMembership.ID.String(),
			NewRole:                  constants.OrgRoleMember,
			TokenScope:               constants.TokenScopeUpdateOrganizationMembership,
		})
		require.NoError(t, err)
		requireOrgMembershipRole(t, suite, alreadyMemberMembership.ID, constants.OrgRoleMember)
	})
}

func Test_RemoveOrganizationMember_ValidationAndFailure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	peerMemberID := uuid.New()
	orgSuite.CreateMember(t, peerMemberID)
	peerMembership := orgSuite.MemberMembership

	leftMemberID := uuid.New()
	orgSuite.CreateMember(t, leftMemberID)
	leftMembership := orgSuite.MemberMembership
	setOrgMembershipStatus(t, suite, leftMembership.ID, constants.OrgMemStatusLeft)

	testCases := []struct {
		name          string
		ctx           context.Context
		orgMemID      string
		expectedError string
	}{
		{
			name:          "invalid_membership_id_returns_validation_error",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      "invalid_id",
			expectedError: string(apperror.CodeValidation),
		},
		{
			name:          "missing_user_metadata_returns_internal_error",
			ctx:           suite.Ctx,
			orgMemID:      memberMembership.ID.String(),
			expectedError: string(apperror.CodeInternal),
		},
		{
			name:          "unknown_membership_returns_not_found",
			ctx:           orgMembershipUserCtx(ownerID),
			orgMemID:      uuid.NewString(),
			expectedError: string(apperror.CodeNotFound),
		},
		{
			name:          "admin_cannot_remove_self",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.AdminMembership.ID.String(),
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "admin_cannot_remove_owner",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      orgSuite.OwnerMembership.ID.String(),
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "member_cannot_remove_peer_member",
			ctx:           orgMembershipUserCtx(memberID),
			orgMemID:      peerMembership.ID.String(),
			expectedError: string(apperror.CodePermissionDenied),
		},
		{
			name:          "cannot_remove_left_membership",
			ctx:           orgMembershipUserCtx(adminID),
			orgMemID:      leftMembership.ID.String(),
			expectedError: string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.RemoveOrganizationMember(tc.ctx, &corev1.IDRequest{
				Id: tc.orgMemID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_RemoveOrganizationMember_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMembership := orgSuite.MemberMembership

	ctx := orgMembershipUserCtx(adminID)

	t.Run("admin_removes_member", func(t *testing.T) {
		_, err := service.RemoveOrganizationMember(ctx, &corev1.IDRequest{
			Id: memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusRemoved)
	})

	t.Run("already_removed_membership_is_idempotent", func(t *testing.T) {
		_, err := service.RemoveOrganizationMember(ctx, &corev1.IDRequest{
			Id: memberMembership.ID.String(),
		})
		require.NoError(t, err)
		requireOrgMembershipStatus(t, suite, memberMembership.ID, constants.OrgMemStatusRemoved)
	})
}

func orgMembershipUserCtx(userID uuid.UUID) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		dto.MetaUserIDKey, userID.String(),
	))
}

func setOrgMembershipStatus(t *testing.T, suite *testutil.TestSuite, membershipID uuid.UUID, status string) {
	t.Helper()

	_, err := suite.Q.UpdateOrganizationMembershipStatus(suite.Ctx, db.UpdateOrganizationMembershipStatusParams{
		ID:     membershipID,
		Status: status,
	})
	require.NoError(t, err)
}

func requireOrgMembershipStatus(t *testing.T, suite *testutil.TestSuite, membershipID uuid.UUID, expectedStatus string) {
	t.Helper()

	membership, err := suite.Q.GetOrganizationMembershipWithAllStatuses(suite.Ctx, membershipID)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, membership.Status)
}

func requireOrgMembershipRole(t *testing.T, suite *testutil.TestSuite, membershipID uuid.UUID, expectedRole string) {
	t.Helper()

	membership, err := suite.Q.GetOrganizationMembershipWithAllStatuses(suite.Ctx, membershipID)
	require.NoError(t, err)
	require.Equal(t, expectedRole, membership.Role)
}
