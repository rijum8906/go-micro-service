package user_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/stretchr/testify/require"
)

func Test_GetMyProfile(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	_, p, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to get a token
	suite.SetClientInfoInContext()
	authRes, err := suite.Login(suite.Ctx, &authv1.LoginRequest{Email: email, Password: password})
	require.Nil(t, err)
	require.NotEmpty(t, authRes)

	// Get the profile
	suite.SetUserInfoInContext(authRes.Tokens.AccessToken.Value)
	profile, err := suite.GetMyProfile(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, profile)

	// Match the profile with the expected values
	require.Equal(t, p.ID.String(), profile.Id)
	require.Equal(t, p.UserID.String(), profile.UserId)
	require.Equal(t, p.FirstName, profile.FirstName)
	require.Equal(t, p.LastName, profile.LastName)
	require.Equal(t, p.AvatarUrl, profile.AvatarUrl)

}

func Test_UpdateProfileName(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	_, p, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to get a token
	suite.SetClientInfoInContext()
	authRes, err := suite.Login(suite.Ctx, &authv1.LoginRequest{Email: email, Password: password})
	require.Nil(t, err)
	require.NotEmpty(t, authRes)
	suite.SetUserInfoInContext(authRes.Tokens.AccessToken.Value)

	// Update the profile name (only first name)
	newFirstName := testutils.GenerateRandomString(10)
	res, err := suite.UserService.UpdateProfileName(suite.Ctx, &userv1.UpdateProfileNameRequest{
		ProfileId: p.ID.String(),
		FirstName: newFirstName,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Get the updated profile
	profile, err := suite.GetMyProfile(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, newFirstName, profile.FirstName)
	require.Equal(t, p.LastName, profile.LastName)
	require.Equal(t, p.AvatarUrl, profile.AvatarUrl)

	// Update the profile name (both first and last name)
	newFirstName = testutils.GenerateRandomString(10)
	newLastName := testutils.GenerateRandomString(10)
	res, err = suite.UserService.UpdateProfileName(suite.Ctx, &userv1.UpdateProfileNameRequest{
		ProfileId: p.ID.String(),
		FirstName: newFirstName,
		LastName:  newLastName,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Get the updated profile
	profile, err = suite.GetMyProfile(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, newFirstName, profile.FirstName)
	require.Equal(t, newLastName, profile.LastName)
	require.Equal(t, p.AvatarUrl, profile.AvatarUrl)
}

func Test_UpdateProfileAvatar(t *testing.T) {
	suite := user.NewTestSuite()
	defer suite.TearDownSuite()

	// Create a user
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)
	_, p, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

	// Login to get a token
	suite.SetClientInfoInContext()
	authRes, err := suite.Login(suite.Ctx, &authv1.LoginRequest{Email: email, Password: password})
	require.Nil(t, err)
	require.NotEmpty(t, authRes)
	suite.SetUserInfoInContext(authRes.Tokens.AccessToken.Value)

	// Update the profile avatar
	avatarUrl := testutils.GenerateRandomString(10)
	res, err := suite.UserService.UpdateProfileAvatarURL(suite.Ctx, &userv1.UpdateProfileAvatarURLRequest{
		ProfileId: p.ID.String(),
		AvatarUrl: avatarUrl,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Get the updated profile
	profile, err := suite.GetMyProfile(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, p.FirstName, profile.FirstName)
	require.Equal(t, p.LastName, profile.LastName)
	require.Equal(t, avatarUrl, profile.AvatarUrl)
}
