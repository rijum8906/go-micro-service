package orgmembership_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	coremetadata "github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/service_test_utils"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// =============================================================================
// SendInvitation Tests
// =============================================================================

func Test_SendInvitation_Validation_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	tests := []struct {
		name string
		req  *org_membershipv1.SendInvitationRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "blank organization id",
			req: &org_membershipv1.SendInvitationRequest{
				OrganizationId: "",
				Email:          "user@example.com",
				Role:           "member",
			},
		},
		{
			name: "malformed organization id",
			req: &org_membershipv1.SendInvitationRequest{
				OrganizationId: "malformed-uuid",
				Email:          "user@example.com",
				Role:           "member",
			},
		},
		{
			name: "blank email",
			req: &org_membershipv1.SendInvitationRequest{
				OrganizationId: uuid.NewString(),
				Email:          "",
				Role:           "member",
			},
		},
		{
			name: "invalid email format",
			req: &org_membershipv1.SendInvitationRequest{
				OrganizationId: uuid.NewString(),
				Email:          "not-an-email",
				Role:           "member",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.SendInvitation(suite.Ctx, tt.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), string(apperror.CodeValidation))
		})
	}
}

func Test_SendInvitation_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
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
			servicetestutils.MockUserServiceClient.On("GetMySelf", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
				Id: uuid.NewString(),
			}, nil)
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
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)
	internEmail := testutils.GenerateRandomEmail()

	ctx := metadata.NewIncomingContext(suite.Ctx, metadata.Pairs(
		dto.MetaUserIDKey, ownerID.String(),
	))

	servicetestutils.MockUserServiceClient.On("CheckEmailExists", ctx, &corev1.EmailRequest{Email: internEmail}).Return(&userv1.CheckExistsResponse{Exists: true}, nil)
	servicetestutils.MockUserServiceClient.On("GetMySelf", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
		Id:    uuid.NewString(),
		Email: internEmail,
	}, nil)

	_, err := service.SendInvitation(ctx, &org_membershipv1.SendInvitationRequest{
		Email:          internEmail,
		Role:           "member",
		OrganizationId: orgSuite.Org.ID.String(),
	})
	require.NoError(t, err)
}

// =============================================================================
// AcceptInvitation Tests
// =============================================================================

func Test_AcceptInvitation_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	// Create an org
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Send and invitation
	email1 := testutils.GenerateRandomEmail()
	invitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email1,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create invalid invitations
	email2 := testutils.GenerateRandomEmail()
	invalidaInvitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email2,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	})
	require.NoError(t, err)
	email3 := testutils.GenerateRandomEmail()
	invalidaInvitation2, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email3,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	})
	require.NoError(t, err)
	_, err = suite.Q.AcceptOrganizationInvitation(suite.Ctx, db.AcceptOrganizationInvitationParams{
		ID:          invalidaInvitation2.ID,
		RespondedBy: uuid.New(),
	})
	require.NoError(t, err)

	testCases := []struct {
		name                string
		setupContext        func() context.Context
		invitationTokenHash string
		expectedUserEmail   string
		organizationID      string
		expectedError       string
	}{
		{
			name:                "without_contexted_user_info_returns_error",
			setupContext:        func() context.Context { return suite.Ctx },
			organizationID:      uuid.New().String(),
			invitationTokenHash: invitation.TokenHash,
			expectedError:       string(apperror.CodeInternal),
		},
		{
			name: "expired_invitation_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email2,
			organizationID:      orgSuite.Org.ID.String(),
			invitationTokenHash: invalidaInvitation.TokenHash,
			expectedError:       string(apperror.CodeNotFound),
		},
		{
			name: "accepted_invitation_won't_be_accepted_twice_returns_not_found_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email3,
			organizationID:      orgSuite.Org.ID.String(),
			invitationTokenHash: invalidaInvitation2.TokenHash,
			expectedError:       string(apperror.CodeNotFound),
		},
		{
			name: "email_not_matches_returns_permission_denied_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   testutils.GenerateRandomEmail(),
			organizationID:      orgSuite.Org.ID.String(),
			invitationTokenHash: invitation.TokenHash,
			expectedError:       string(apperror.CodePermissionDenied),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, ok := coremetadata.GetUserInfoFromIncomingContext(ctx)
			if ok {
				servicetestutils.MockUserServiceClient.On("GetUser", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
					Email: tc.expectedUserEmail,
				}, nil)
			}

			_, err := service.AcceptInvitation(ctx, &corev1.TokenHashRequest{
				TokenHash: tc.invitationTokenHash,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_AcceptInvitation_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Create invitation
	email := testutils.GenerateRandomEmail()
	invitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24),
			Valid: true,
		},
	})
	require.NoError(t, err)

	userID := uuid.New()
	ctx := metadata.NewIncomingContext(suite.Ctx, metadata.Pairs(
		dto.MetaUserIDKey, userID.String(),
	))

	servicetestutils.MockUserServiceClient.On("GetUser", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
		Email: email,
	}, nil)

	res, err := service.AcceptInvitation(ctx, &corev1.TokenHashRequest{
		TokenHash: invitation.TokenHash,
	})
	require.NoError(t, err)
	require.True(t, res.Success)

	// Check the actual invitation
	invitation, err = suite.Q.GetOrganizationInvitationWithAllStatus(suite.Ctx, invitation.ID)
	require.NoError(t, err)
	require.Equal(t, invitation.RespondedAt.Valid, true)
	require.Equal(t, invitation.RespondedByUserID.String(), userID.String())
}

// =============================================================================
// DeclineInvitation Tests
// =============================================================================

func Test_DeclineInvitation_Integration_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	// Create an org
	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Send an invitation
	email1 := testutils.GenerateRandomEmail()
	invitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email1,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24),
			Valid: true,
		},
	})
	require.NoError(t, err)

	// Create expired invitation
	email2 := testutils.GenerateRandomEmail()
	expiredInvitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email2,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(-time.Hour), // Expired
			Valid: true,
		},
	})
	require.NoError(t, err)

	// Create already accepted invitation
	email3 := testutils.GenerateRandomEmail()
	acceptedInvitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email3,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	})
	require.NoError(t, err)

	// Accept this invitation first
	_, err = suite.Q.AcceptOrganizationInvitation(suite.Ctx, db.AcceptOrganizationInvitationParams{
		ID:          acceptedInvitation.ID,
		RespondedBy: uuid.New(),
	})
	require.NoError(t, err)

	// Create already declined invitation
	email4 := testutils.GenerateRandomEmail()
	declinedInvitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email4,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	})
	require.NoError(t, err)

	// Decline this invitation first
	_, err = suite.Q.DeclineOrganizationInvitation(suite.Ctx, db.DeclineOrganizationInvitationParams{
		ID:          declinedInvitation.ID,
		RespondedBy: uuid.New(),
	})
	require.NoError(t, err)

	testCases := []struct {
		name                string
		setupContext        func() context.Context
		invitationTokenHash string
		expectedUserEmail   string
		expectedError       string
	}{
		{
			name: "without_contexted_user_info_returns_error",
			setupContext: func() context.Context {
				return suite.Ctx
			},
			invitationTokenHash: invitation.TokenHash,
			expectedError:       string(apperror.CodeInternal),
		},
		{
			name: "expired_invitation_returns_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email2,
			invitationTokenHash: expiredInvitation.TokenHash,
			expectedError:       string(apperror.CodeNotFound),
		},
		{
			name: "already_accepted_invitation_returns_not_found_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email3,
			invitationTokenHash: acceptedInvitation.TokenHash,
			expectedError:       string(apperror.CodeNotFound),
		},
		{
			name: "already_declined_invitation_returns_not_found_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email4,
			invitationTokenHash: declinedInvitation.TokenHash,
			expectedError:       string(apperror.CodeNotFound),
		},
		{
			name: "email_not_matches_returns_permission_denied_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   testutils.GenerateRandomEmail(),
			invitationTokenHash: invitation.TokenHash,
			expectedError:       string(apperror.CodePermissionDenied),
		},
		{
			name: "nil_request_returns_validation_error",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					dto.MetaUserIDKey, uuid.New().String(),
				))
			},
			expectedUserEmail:   email1,
			invitationTokenHash: "", // Empty token hash
			expectedError:       string(apperror.CodeValidation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			_, ok := coremetadata.GetUserInfoFromIncomingContext(ctx)
			if ok {
				servicetestutils.MockUserServiceClient.On("GetUser", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
					Email: tc.expectedUserEmail,
				}, nil)
			}

			_, err := service.DeclineInvitation(ctx, &corev1.TokenHashRequest{
				TokenHash: tc.invitationTokenHash,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func Test_DeclineInvitation_Integration_Success(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	// Create invitation
	email := testutils.GenerateRandomEmail()
	invitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24),
			Valid: true,
		},
	})
	require.NoError(t, err)

	userID := uuid.New()
	ctx := metadata.NewIncomingContext(suite.Ctx, metadata.Pairs(
		dto.MetaUserIDKey, userID.String(),
	))

	servicetestutils.MockUserServiceClient.On("GetUser", ctx, &corev1.EmptyRequest{}).Return(&modelsv1.User{
		Email: email,
	}, nil)

	res, err := service.DeclineInvitation(ctx, &corev1.TokenHashRequest{
		TokenHash: invitation.TokenHash,
	})
	require.NoError(t, err)
	require.True(t, res.Success)

	// Verify the invitation was marked as declined
	declinedInvitation, err := suite.Q.GetOrganizationInvitationWithAllStatus(suite.Ctx, invitation.ID)
	require.NoError(t, err)
	require.True(t, declinedInvitation.RespondedAt.Valid, "RespondedAt should be set")
	require.Equal(t, userID.String(), declinedInvitation.RespondedByUserID.String(), "RespondedBy should match the user who declined")

	// Verify the invitation status is no longer "pending"
	// Note: This depends on your DeclineOrganizationInvitation implementation
	// You may need to check if status was updated to "declined" or similar
	require.NotEqual(t, "pending", declinedInvitation.Status, "Invitation status should not be pending")

	// Verify no membership was created (this is the key difference from AcceptInvitation)
	_, err = suite.Q.GetOrganizationMembershipByOrgIDAndUserID(suite.Ctx, db.GetOrganizationMembershipByOrgIDAndUserIDParams{
		UserID:         userID,
		OrganizationID: orgSuite.Org.ID,
	})
	require.Error(t, err, "Membership should not exist for declined invitation")
	require.True(t, errors.Is(err, sql.ErrNoRows), "Should return no rows error for non-existent membership")
}

func Test_DeclineInvitation_Integration_UserServiceError(t *testing.T) {
	suite := testutil.NewTestSuite(t, fgaClient)
	service := suite.Service

	ownerID := uuid.New()
	orgSuite := suite.CreateOrg(t, ownerID)
	orgSuite.CreateOwner(t, ownerID)

	email := testutils.GenerateRandomEmail()
	invitation, err := suite.Q.CreateOrganizationInvitation(suite.Ctx, db.CreateOrganizationInvitationParams{
		Email:          email,
		Role:           "intern",
		TokenHash:      testutils.GenerateRandomString(32),
		OrganizationID: orgSuite.Org.ID,
		InvitedByMemID: orgSuite.OwnerMembership.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24),
			Valid: true,
		},
	})
	require.NoError(t, err)

	userID := uuid.New()
	ctx := metadata.NewIncomingContext(suite.Ctx, metadata.Pairs(
		dto.MetaUserIDKey, userID.String(),
	))

	// Mock user service error
	expectedErr := apperror.ErrThirdParty.WithMessage("user service unavailable")
	servicetestutils.MockUserServiceClient.On("GetUser", ctx, &corev1.EmptyRequest{}).Return(nil, expectedErr)

	_, err = service.DeclineInvitation(ctx, &corev1.TokenHashRequest{
		TokenHash: invitation.TokenHash,
	})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
