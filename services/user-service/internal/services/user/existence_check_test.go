package user_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/stretchr/testify/require"
)

func Test_CheckEmailExists(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	_, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// This email should exist
	exists, err := suite.UserService.CheckEmailExists(suite.Ctx, &corev1.EmailRequest{Email: email})
	require.NoError(t, err)
	require.True(t, exists.Exists)

	// Random email should not exist
	nonExists, err := suite.UserService.CheckEmailExists(suite.Ctx, &corev1.EmailRequest{Email: testutils.GenerateRandomEmail()})
	require.NoError(t, err)
	require.False(t, nonExists.Exists)
}

func Test_CheckExists(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Check that the user exists
	exists, err := suite.UserService.CheckExists(suite.Ctx, &corev1.IDRequest{Id: user.ID.String()})
	require.NoError(t, err)
	require.True(t, exists.Exists)

	// Check that a random ID does not exist
	nonExists, err := suite.UserService.CheckExists(suite.Ctx, &corev1.IDRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	require.False(t, nonExists.Exists)
}
