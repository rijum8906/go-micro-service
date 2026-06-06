package session_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/stretchr/testify/require"
)

func Test_GetSessions(t *testing.T) {
	suite := session.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to multiple devices
	authResponses := loginToMultipleDevices(t, suite, email, password, 5)
	authResponse := authResponses[len(authResponses)-1]

	// Get the sessions
	suite.SetUserInfoInContext(authResponse.Tokens.AccessToken.Value)
	sessions, err := suite.SessionService.GetSessions(suite.Ctx, &corev1.PaginationRequest{
		Page:  1,
		Limit: 5,
	})
	require.Nil(t, err)
	require.Len(t, sessions.Sessions, 3)

	// All the sessions should belong to the user
	for _, session := range sessions.Sessions {
		require.Equal(t, user.ID.String(), session.UserId)
	}
}

func Test_GetActiveSession(t *testing.T) {
	suite := session.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to multiple devices
	authResponses := loginToMultipleDevices(t, suite, email, password, 5)
	authResponse := authResponses[len(authResponses)-1]

	// The context holds the last client info
	// the authResponse contains the refresh token we can use to logout the last device
	suite.SetUserInfoInContext(authResponse.Tokens.AccessToken.Value) // logout needs user to be authenticated
	successRes, err := suite.AuthService.Logout(suite.Ctx, &authv1.LogoutRequest{
		RefreshToken: authResponse.Tokens.RefreshToken.Value,
	})
	require.Nil(t, err)
	require.True(t, successRes.Success)

	// Get the sessions
	sessions, err := suite.SessionService.GetActiveSessions(suite.Ctx, &corev1.PaginationRequest{
		Page:  1,
		Limit: 10,
	})
	require.Nil(t, err)
	require.Len(t, sessions.Sessions, 4)

	// All the sessions should belong to the user
	for _, session := range sessions.Sessions {
		require.Equal(t, user.ID.String(), session.UserId)
	}
}

func Test_GetCurrentSession(t *testing.T) {
	suite := session.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	_, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to multiple devices
	authResponses := loginToMultipleDevices(t, suite, email, password, 5)
	authResponse := authResponses[len(authResponses)-1]

	// The last login should be the current session
	suite.SetUserInfoInContext(authResponse.Tokens.AccessToken.Value)
	res, err := suite.SessionService.GetCurrentSession(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Get the current session again via dbq
	res2, err := suite.AuthService.DBQ.GetSessionByRefreshTokenHash(suite.Ctx, authResponse.Tokens.RefreshToken.Value)
	require.NoError(t, err)
	require.NotNil(t, res2)

	// Match the two sessions
	require.Equal(t, res.DeviceId, res2.DeviceID)
	require.Equal(t, res.Id, res2.ID.String())
	require.Equal(t, res.UserId, res2.UserID.String())
}
