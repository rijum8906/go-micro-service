package user_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/stretchr/testify/require"
)

func Test_GetMyself(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to get a token
	suite.SetClientInfoInContext()
	authRes, err := suite.Login(suite.Ctx, &authv1.LoginRequest{Email: email, Password: password})
	require.Nil(t, err)
	require.NotEmpty(t, authRes)

	// Get my self
	suite.SetUserInfoInContext(authRes.Tokens.AccessToken.Value)
	userRes, err := suite.UserService.GetMySelf(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, userRes)

	// Match the user
	require.Equal(t, user.ID.String(), userRes.Id)
	require.Equal(t, user.Email, userRes.Email)
}

func Test_GetUser(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Get the user
	userRes, err := suite.UserService.GetUser(suite.Ctx, &corev1.IDRequest{Id: user.ID.String()})
	require.NoError(t, err)
	require.NotNil(t, userRes)

	// Match the user
	require.Equal(t, user.ID.String(), userRes.Id)
	require.Equal(t, user.Email, userRes.Email)
}
