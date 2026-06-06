package auth_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// Preconditions:
//   - Required a valid session ID in user info in ctx
func Test_Logout(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// user credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Attach client info in ctx
	suite.SetClientInfoInContext()
	ctx := suite.Ctx

	// Create a user
	user, _, appErr := suite.CreateUser(t, email, password)
	require.Nil(t, appErr)

	// Login
	authRes, appErr := suite.AuthService.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes.Status.String())

	// Retrive the session ID from access token
	claims, appErr := suite.AuthService.TokenManager.ValidateAuthToken(ctx, authRes.Tokens.AccessToken.Value)
	require.Nil(t, appErr)

	// Attach user info in ctx
	suite.SetUserInfoInContext(user)
	md, _ := metadata.FromIncomingContext(ctx)
	md.Set(dto.MetaSessionIDKey, claims.SessionID)
	ctx = metadata.NewIncomingContext(ctx, md)

	// Logout
	res, err := suite.AuthService.Logout(ctx, &authv1.LogoutRequest{})
	require.Nil(t, err)
	require.True(t, res.Success)
}

// Test_LogoutAllDevices tests the logout all devices functionality
// Preconditions:
// - Two authenticated sessions for the same user with different client info
func Test_LogoutAllDevices(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	// user credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Create a user
	user, _, appErr := suite.CreateUser(t, email, password)
	require.Nil(t, appErr)

	// Login 1
	suite.SetClientInfoInContext() // Set client info in ctx1
	authRes1, appErr := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes1.Status.String())

	// Login 2
	suite.SetClientInfoInContext() // Set client info in ctx2
	authRes2, appErr := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes2.Status.String())

	// Logout all devices
	suite.SetUserInfoInContext(user)
	res, err := suite.AuthService.LogoutAllDevices(suite.Ctx, &corev1.EmptyRequest{})
	require.Nil(t, err)
	require.True(t, res.Success)

	// Should not be able to validate the access token after logout
	_, appErr = suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, authRes1.Tokens.AccessToken.Value)
	require.NotNil(t, appErr)

	// Should not be able to validate the access token after logout
	_, appErr = suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, authRes2.Tokens.AccessToken.Value)
	require.NotNil(t, appErr)
}
