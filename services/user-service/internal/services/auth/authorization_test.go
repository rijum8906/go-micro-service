package auth_test

import (
	"context"
	"testing"

	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// Preconditions:
//   - Required a valid session ID in user info in ctx
func Test_Logout(t *testing.T) {
	ctx := context.Background()

	// user credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Attach client info in ctx
	ctx = setClientInfoInCtx(ctx)

	// Create a user
	user, _, appErr := createUser(ctx, email, password)
	require.Nil(t, appErr)

	// Login
	authRes, appErr := authService.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes.Status.String())

	// Retrive the session ID from access token
	claims, appErr := authService.TokenManager.ValidateAuthToken(ctx, authRes.Tokens.AccessToken.Value)
	require.Nil(t, appErr)

	// Attach user info in ctx
	ctx = setUserInfoInCtx(ctx, user)
	md, _ := metadata.FromIncomingContext(ctx)
	md.Set(dto.MetaSessionIDKey, claims.SessionID)
	ctx = metadata.NewIncomingContext(ctx, md)

	// Logout
	res, err := authService.Logout(ctx, &authv1.LogoutRequest{})
	require.Nil(t, err)
	require.True(t, res.Success)
}

// Test_LogoutAllDevices tests the logout all devices functionality
// Preconditions:
// - Two authenticated sessions for the same user with different client info
func Test_LogoutAllDevices(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()

	// user credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Create a user
	user, _, appErr := createUser(ctx1, email, password)
	require.Nil(t, appErr)

	// Login 1
	ctx1 = setClientInfoInCtx(ctx1) // Set client info in ctx1
	ctx1 = setUserInfoInCtx(ctx1, user)
	authRes1, appErr := authService.Login(ctx1, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes1.Status.String())

	// Login 2
	ctx2 = setClientInfoInCtx(ctx2) // Set client info in ctx2
	ctx2 = setUserInfoInCtx(ctx2, user)
	authRes2, appErr := authService.Login(ctx2, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.Nil(t, appErr)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_SUCCESS.String(), authRes2.Status.String())

	// Logout all devices
	res, err := authService.LogoutAllDevices(ctx2, &corev1.EmptyRequest{})
	require.Nil(t, err)
	require.True(t, res.Success)

	// Should not be able to validate the access token after logout
	_, appErr = authService.TokenManager.ValidateAuthToken(ctx1, authRes1.Tokens.AccessToken.Value)
	require.NotNil(t, appErr)

	// Should not be able to validate the access token after logout
	_, appErr = authService.TokenManager.ValidateAuthToken(ctx2, authRes2.Tokens.AccessToken.Value)
	require.NotNil(t, appErr)

}
