package auth_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	mock_broker "github.com/rijum8906/relay/packages/core/broker/mocks"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	mock_token "github.com/rijum8906/relay/packages/core/token/mocks"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TODO: validation test for this function

// Preconditions:
//   - Must create a user
//   - Mock publisher with (jobs.JobUserRequestedEmailVerification, dto.EmailVerificationDTO) args (nil)
//   - Requires client info be attached to the context (not implemented yet)
func Test_RequestEmailVerification(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Attach client info to context
	suite.SetClientInfoInContext()

	// Mock publisher
	mockPublisher, ok := suite.AuthService.BrokerPublisher.(*mock_broker.MockPublisher)
	require.True(t, ok)
	mockPublisher.On("Publish", jobs.JobUserRequestedEmailVerification, mock.Anything).Return(nil)

	// Call the function under test
	res, err := suite.AuthService.RequestEmailVerification(suite.Ctx, &authv1.RequestEmailVerificationRequest{
		Email: user.Email,
	})
	require.Nil(t, err)
	require.True(t, res.Success)
}

func Test_RequestPasswordReset(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Mock publisher
	mockPublisher, ok := suite.AuthService.BrokerPublisher.(*mock_broker.MockPublisher)
	require.True(t, ok)
	mockPublisher.On("Publish", jobs.JobUserRequestedPasswordReset, mock.Anything).Return(nil)

	// Call the function under test
	res, err := suite.AuthService.RequestPasswordReset(suite.Ctx, &authv1.RequestPasswordResetRequest{
		Email: user.Email,
	})
	require.Nil(t, err)
	require.True(t, res.Success)
}

// Preconditions:
//   - Must create a user
//   - Mock tokenManager to accept the verification token and process verification
func Test_VerifyEmail(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	scopedToken := testutils.GenerateRandomString(64) // Generate a random string which will act as the scoped token

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Mock tokenManager
	mockTokenManager := mock_token.NewMockTokenManager(t)
	mockTokenManager.On("ValidateScopedToken", suite.Ctx, scopedToken).Return(&token.Claims{
		Scope:     constants.TokenScopeVerifyEmail,
		SessionID: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID.String(),
		},
	}, nil)
	suite.AuthService.TokenManager = mockTokenManager

	// Call the function under test
	res, err := suite.AuthService.VerifyEmail(suite.Ctx, &authv1.VerifyEmailRequest{
		ScopedToken: scopedToken,
	})
	require.NoError(t, err)
	require.True(t, res.Success)

	// We should see the user as IsEmailVerified = true
	updatedUser, err := suite.AuthService.DBQ.GetUser(suite.Ctx, user.ID)
	require.NoError(t, err)
	require.True(t, updatedUser.IsEmailVerified)
}

func Test_ResetPassword(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	scopedToken := testutils.GenerateRandomString(64) // Generate a random string which will act as the scoped token

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Mock tokenManager
	mockTokenManager := mock_token.NewMockTokenManager(t)
	mockTokenManager.On("ValidateScopedToken", suite.Ctx, scopedToken).Return(&token.Claims{
		Scope:     constants.TokenScopeResetPassword,
		SessionID: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID.String(),
		},
	}, nil)
	suite.AuthService.TokenManager = mockTokenManager

	// Call the function under test
	newPassword := testutils.GenerateRandomString(10)
	res, err := suite.AuthService.ResetPassword(suite.Ctx, &authv1.ResetPasswordRequest{
		ScopedToken: scopedToken,
		NewPassword: newPassword,
	})
	require.NoError(t, err)
	require.True(t, res.Success)

	// We should match the updated password
	updatedUser, err := suite.AuthService.DBQ.GetUser(suite.Ctx, user.ID)
	require.NoError(t, err)

	isPasswordMatched := suite.AuthService.HashService.Verify(updatedUser.PasswordHash.String, newPassword)
	require.True(t, isPasswordMatched)
}
