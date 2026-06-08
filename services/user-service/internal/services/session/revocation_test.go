package session_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/stretchr/testify/require"
)

func loginToMultipleDevices(t *testing.T, suite *session.TestSuite, email, password string, count int) []*authv1.AuthResponse {
	t.Helper()

	authResponses := make([]*authv1.AuthResponse, 0, count)
	for range count {
		suite.SetClientInfoInContext()
		authResponse, err := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
			Email:    email,
			Password: password,
		})
		require.NoError(t, err)
		require.NotNil(t, authResponse)
		authResponses = append(authResponses, authResponse)
	}

	return authResponses
}

func requireTokenRevoked(t *testing.T, suite *session.TestSuite, accessToken string) {
	t.Helper()

	claims, appErr := suite.SessionService.TokenManager.ValidateAuthToken(suite.Ctx, accessToken)
	require.NotNil(t, appErr)
	require.Nil(t, claims)
}

func getActiveSessions(t *testing.T, suite *session.TestSuite, user db.User) []db.Session {
	t.Helper()

	activeSessions, err := suite.SessionService.DBQ.GetActiveSessionsByUserID(suite.Ctx, db.GetActiveSessionsByUserIDParams{
		UserID: user.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)

	return activeSessions
}

func Test_RevokeSession(t *testing.T) {
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

	// The context holds the last client info
	// the authResponse contains the refresh token we can use to logout the last device
	suite.SetUserInfoInContext(authResponse.Tokens.AccessToken.Value) // logout needs user to be authenticated
	successRes, err := suite.SessionService.RevokeSession(suite.Ctx, &sessionv1.RevokeSessionRequest{
		TokenToRevoke: authResponse.Tokens.AccessToken.Value,
	})
	require.Nil(t, err)
	require.True(t, successRes.Success)

	// Validate the access token should give me error
	requireTokenRevoked(t, suite, authResponse.Tokens.AccessToken.Value)

	// Session in the database should be marked as revoked
	session, err := suite.SessionService.DBQ.GetRevokedSessionByRefreshTokenHash(suite.Ctx, authResponse.Tokens.RefreshToken.Value)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.True(t, session.IsRevoked)
}

func Test_RevokeAllSessions(t *testing.T) {
	suite := session.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to multiple devices
	authResponses := loginToMultipleDevices(t, suite, email, password, 3)

	// Revoke all sessions
	suite.SetUserInfoInContext(authResponses[len(authResponses)-1].Tokens.AccessToken.Value)
	successRes, err := suite.SessionService.RevokeAllSessions(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.True(t, successRes.Success)

	// All access tokens should be revoked
	for _, authResponse := range authResponses {
		requireTokenRevoked(t, suite, authResponse.Tokens.AccessToken.Value)
	}

	// There should be no active sessions left for the user
	activeSessions := getActiveSessions(t, suite, *user)
	require.Empty(t, activeSessions)
}

func Test_RevokeOtherSessions(t *testing.T) {
	suite := session.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to multiple devices
	authResponses := loginToMultipleDevices(t, suite, email, password, 3)
	currentAuthResponse := authResponses[len(authResponses)-1]
	currentClaims, appErr := suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, currentAuthResponse.Tokens.AccessToken.Value)
	require.Nil(t, appErr)

	// Revoke every session except the current one, issuing fresh current-session tokens
	suite.SetUserInfoInContext(currentAuthResponse.Tokens.AccessToken.Value)
	newTokens, err := suite.SessionService.RevokeOtherSessions(suite.Ctx, &sessionv1.RevokeOtherSessionsRequest{})
	require.NoError(t, err)
	require.NotNil(t, newTokens)
	require.NotNil(t, newTokens.AccessToken)
	require.NotNil(t, newTokens.RefreshToken)
	require.NotEmpty(t, newTokens.AccessToken.Value)
	require.NotEmpty(t, newTokens.RefreshToken.Value)
	require.NotEqual(t, currentAuthResponse.Tokens.RefreshToken.Value, newTokens.RefreshToken.Value)

	// The old access tokens should be revoked, including the previous token for the current session.
	for _, authResponse := range authResponses {
		requireTokenRevoked(t, suite, authResponse.Tokens.AccessToken.Value)
	}

	// The new access token should be valid for the same current session.
	newClaims, appErr := suite.SessionService.TokenManager.ValidateAuthToken(suite.Ctx, newTokens.AccessToken.Value)
	require.Nil(t, appErr)
	require.Equal(t, user.ID.String(), newClaims.Subject)
	require.Equal(t, currentClaims.SessionID, newClaims.SessionID)

	// Only the current session should remain active in the database.
	activeSessions := getActiveSessions(t, suite, *user)
	require.Len(t, activeSessions, 1)
	require.Equal(t, currentClaims.SessionID, activeSessions[0].ID.String())

	// The refresh token for the current session should have rotated.
	currentSession, err := suite.SessionService.DBQ.GetSessionByRefreshTokenHash(suite.Ctx, newTokens.RefreshToken.Value)
	require.NoError(t, err)
	require.Equal(t, currentClaims.SessionID, currentSession.ID.String())
	require.False(t, currentSession.IsRevoked)

	// Other sessions should be marked revoked.
	for _, authResponse := range authResponses[:len(authResponses)-1] {
		revokedSession, err := suite.SessionService.DBQ.GetRevokedSessionByRefreshTokenHash(suite.Ctx, authResponse.Tokens.RefreshToken.Value)
		require.NoError(t, err)
		require.True(t, revokedSession.IsRevoked)
	}
}
