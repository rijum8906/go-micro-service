package auth_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	mock_token "github.com/rijum8906/relay/packages/core/token/mocks"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ChangePassword(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a test user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)
	suite.SetUserInfoInContext(user) // Set user info in context for auth token validation

	scopedToken := testutils.GenerateRandomString(10) // A random string acting as a scope token

	// Mock the token manager to pass scope token verification to change password
	mockTokenManager := mock_token.NewMockTokenManager(t)
	mockTokenManager.On("ValidateScopedToken", suite.Ctx, scopedToken).Return(&token.Claims{
		Scope: constants.TokenScopeChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID.String(),
		},
	}, nil)
	suite.AuthService.TokenManager = mockTokenManager

	// Call the change password function

	newPassword := testutils.GenerateRandomString(10)
	res, err := suite.AuthService.ChangePassword(suite.Ctx, &authv1.ChangePasswordRequest{
		NewPassword: newPassword,
		ScopedToken: scopedToken,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Check that the password has been updated
	updatedUser, err := suite.AuthService.DBQ.GetUser(suite.Ctx, user.ID)
	require.NoError(t, err)
	isPasswordMatched := suite.AuthService.HashService.Verify(updatedUser.PasswordHash.String, newPassword)
	assert.True(t, isPasswordMatched)

}
