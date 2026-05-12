package orgmembership_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/token"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/constants"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// =============================================================================
// ChangeOrganizationMembershipStatus Tests (Additional comprehensive tests)
// =============================================================================

func Test_ChangeOrganizationMembershipStatus_Validation(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	testCases := []struct {
		name              string
		setupContext      func() context.Context
		req               *org_membershipv1.ChangeOrgMembershipStatusReq
		expectedErrorCode string
	}{
		{
			name: "invalid_membership_id",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			req: &org_membershipv1.ChangeOrgMembershipStatusReq{
				OrganizationMembershipId: "invalid_membership_id",
			},
			expectedErrorCode: string(apperror.CodeValidation),
		},
		{
			name: "invalid_token_scope",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			req: &org_membershipv1.ChangeOrgMembershipStatusReq{
				OrganizationMembershipId: uuid.NewString(),
				TokenScope:               "invalid_scope",
			},
			expectedErrorCode: string(apperror.CodeValidation),
		},
		{
			name: "invalid_status",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			req: &org_membershipv1.ChangeOrgMembershipStatusReq{
				OrganizationMembershipId: uuid.NewString(),
				TokenScope:               string(token.TokenScopeAdmin),
				NewStatus:                "invalid_status",
			},
			expectedErrorCode: string(apperror.CodeValidation),
		},
		{
			name: "wrong_token_scope",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			req: &org_membershipv1.ChangeOrgMembershipStatusReq{
				OrganizationMembershipId: uuid.NewString(),
				TokenScope:               string(token.TokenScopeAdmin),
				NewStatus:                "active",
			},
			expectedErrorCode: string(apperror.CodePermissionDenied),
		},
		{
			name: "without_user_info_containing_context",
			setupContext: func() context.Context {
				return context.Background()
			},
			req: &org_membershipv1.ChangeOrgMembershipStatusReq{
				OrganizationMembershipId: uuid.NewString(),
				TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
				NewStatus:                constants.OrgMemStatusSuspended,
			},
			expectedErrorCode: string(apperror.CodeInternal),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ChangeOrganizationMembershipStatus(tc.setupContext(), tc.req)
			require.NotNil(t, err)
			require.Contains(t, err.Error(), tc.expectedErrorCode)
		})
	}
}

func Test_ChangeOrganizationMembershipStatus_Integration_Validation(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ctx := context.Background()
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	t.Run("owner's status can't be changed", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))
		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: orgSuite.OwnerMembership.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.Error(t, err)
	})

	t.Run("can't set left staus", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))
		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: orgSuite.OwnerMembership.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusLeft,
		})
		require.Error(t, err)
	})
}

func Test_ChangeOrganizationMembershipStatus_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ctx := context.Background()
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	t.Run("actor without permission can't change status", func(t *testing.T) {
		memberID := uuid.New()
		orgSuite.CreateMember(t, memberID)

		memberID2 := uuid.New()
		orgSuite.CreateMember(t, memberID2)
		membership2 := orgSuite.MemberMembership

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, memberID.String(),
		))

		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: membership2.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), string(apperror.CodePermissionDenied))
		require.Contains(t, err.Error(), "you do not have permission to change this membership's status")
	})

	t.Run("after suspending a member he can't view organization", func(t *testing.T) {
		memberID := uuid.New()
		orgSuite.CreateMember(t, memberID)
		merbership := orgSuite.MemberMembership

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))
		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: merbership.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.NoError(t, err)

		res, appErr := suite.TuppleManager.Check(ctx, client.ClientCheckRequest{
			User:     "user:" + memberID.String(),
			Relation: permissions.PermissionCanViewMembers,
			Object:   "organization:" + orgSuite.Org.ID.String(),
		})
		require.Nil(t, appErr)
		require.False(t, *res.Allowed)
	})
}

func Test_ChangeOrganizationMembershipStatus_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ctx := context.Background()
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	adminID := uuid.New()
	orgSuite.CreateAdmin(t, adminID)

	memberID := uuid.New()
	orgSuite.CreateMember(t, memberID)
	memberMem := orgSuite.MemberMembership

	t.Run("owner can change status", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, ownerID.String(),
		))
		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: memberMem.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.NoError(t, err)

		membership, err := suite.Q.GetOrganizationMembership(suite.Ctx, memberMem.ID)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, constants.OrgMemStatusSuspended, membership.Status)
	})

	t.Run("admin can change status", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, adminID.String(),
		))
		_, err := service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: memberMem.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.NoError(t, err)
	})

	t.Run("user with custom role can change status", func(t *testing.T) {
		userID := uuid.New()
		orgSuite.CreateMember(t, userID)
		userMembership := orgSuite.MemberMembership

		appErr := suite.PermissionManager.CreateCustomRole(ctx, userID.String(), orgSuite.Org.ID.String(), "custom_role",
			permissions.PermissionCanChangeMemberStatus)
		require.Nil(t, appErr)
		_, err := suite.Q.UpdateOrganizationMembershipRole(ctx, db.UpdateOrganizationMembershipRoleParams{
			Role: "custom_role",
			ID:   userMembership.ID,
		})
		require.NoError(t, err)

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
			dto.MetaUserIDKey, userID.String(),
		))

		_, err = service.ChangeOrganizationMembershipStatus(ctx, &org_membershipv1.ChangeOrgMembershipStatusReq{
			OrganizationMembershipId: memberMem.ID.String(),
			TokenScope:               string(token.TokenScopeUpdateOrganizationMembership),
			NewStatus:                constants.OrgMemStatusSuspended,
		})
		require.NoError(t, err)

		membership, err := suite.Q.GetOrganizationMembership(suite.Ctx, memberMem.ID)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, constants.OrgMemStatusSuspended, membership.Status)
	})
}
