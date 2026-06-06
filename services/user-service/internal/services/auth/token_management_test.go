package auth_test

import (
	"testing"

	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/require"
)

// Precondition:
//   - Login function must be tested and work correctly
func Test_GenerateScopedToken(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Login the user
	suite.SetClientInfoInContext()
	authRes, err := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotNil(t, authRes)

	// try to generate a scoped token
	suite.SetUserInfoInContext(user) // set user info in context to simulate authenticated request
	tokenRes, err := suite.AuthService.GenerateScopedToken(suite.Ctx, &authv1.GenerateScopedTokenRequest{
		AuthMethod: string(coreconstants.AuthMethodPassword),
		Scope:      constants.TokenScopeChangePassword,
		AuthValue:  password,
	})
	require.NoError(t, err)
	require.NotNil(t, tokenRes)

	// Validate the token with the new password
	claims, appErr := suite.AuthService.TokenManager.ValidateScopedToken(suite.Ctx, tokenRes.ScopedToken)
	require.Nil(t, appErr)
	require.NotNil(t, claims)
	require.Equal(t, claims.Scope, constants.TokenScopeChangePassword)
	require.Equal(t, claims.Subject, user.ID.String())
}

// TODO: add test cases for validation and edge cases

func Test_RefreshToken(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Login the user
	suite.SetClientInfoInContext()
	authRes, err := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotNil(t, authRes)

	// Manually invalidate the access token
	claims, appErr := suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, authRes.Tokens.AccessToken.Value)
	require.Nil(t, appErr)
	require.Equal(t, user.ID.String(), claims.Subject)
	require.NotEmpty(t, claims.SessionID)
	appErr = suite.AuthService.TokenManager.RevokeAuthToken(suite.Ctx, claims.ID, user.ID.String())
	require.Nil(t, appErr)

	// Refresh the token
	refreshRes, err := suite.AuthService.RefreshToken(suite.Ctx, &authv1.RefreshAccessTokenRequest{
		RefreshToken: authRes.Tokens.RefreshToken.Value,
	})
	require.NoError(t, err)
	require.NotNil(t, refreshRes)
	require.NotNil(t, refreshRes.AccessToken)
	require.NotEmpty(t, refreshRes.AccessToken.Value)
	require.NotEmpty(t, refreshRes.RefreshToken)
	require.NotEqual(t, authRes.Tokens.RefreshToken.Value, refreshRes.RefreshToken)

	claims, appErr = suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, refreshRes.AccessToken.Value)
	require.Nil(t, appErr)
	require.Equal(t, user.ID.String(), claims.Subject)
	require.NotEmpty(t, claims.SessionID)
}
