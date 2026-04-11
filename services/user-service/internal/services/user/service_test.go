package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/services/user"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func createTestAuthService() (auth.AuthService, *utils.Repos, *utils.ServiceUtils) {
	repos := utils.NewTestRepos()
	serviceUtils := utils.NewTestServiceUtils()
	config := testutils.NewTestEnv()
	service, appErr := auth.NewAuthService(repos, serviceUtils, config)
	if appErr != nil {
		panic(appErr)
	}
	return service, repos, serviceUtils
}

func createTestUserService() (user.UserService, *utils.Repos, *utils.ServiceUtils) {
	repos := utils.NewTestRepos()
	serviceUtils := utils.NewTestServiceUtils()
	config := testutils.NewTestEnv()
	service, appErr := user.NewUserService(repos, serviceUtils, config)
	if appErr != nil {
		panic(appErr)
	}
	return service, repos, serviceUtils
}

func MustCreateUserAndProfile(service auth.AuthService, password string) *authv1.AuthResponse {
	u, appErr := service.Register(context.Background(), &authv1.RegisterRequest{
		Email:     testutils.GenerateRandomEmail(),
		Password:  password,
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
	}, testutils.GenerateClientInfo())
	if appErr != nil {
		fmt.Println("Registration failed: ", appErr.Code)
		panic(appErr)
	}

	return u
}

func TestUserService_GetProfile(t *testing.T) {
	userService, _, serviceUtils := createTestUserService()
	authService, _, _ := createTestAuthService()

	registerRes := MustCreateUserAndProfile(authService, "test1234")

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	getProfile, appErr := userService.GetProfile(context.Background(), &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if getProfile == nil {
		t.Fatal("expected profile, got nil")
	}
	if getProfile.UserId != registerRes.User.Id {
		t.Fatalf("expected user id %v, got %v", registerRes.User.Id, getProfile.UserId)
	}
	if getProfile.Id != registerRes.Profile.Id {
		t.Fatalf("expected profile id %v, got %v", registerRes.Profile.Id, getProfile.Id)
	}
	if getProfile.AvatarUrl != registerRes.Profile.AvatarUrl {
		t.Fatalf("expected avatar url %v, got %v", registerRes.Profile.AvatarUrl, getProfile.AvatarUrl)
	}
}

func TestUserService_UpdateProfileName(t *testing.T) {
	userService, _, serviceUtils := createTestUserService()
	authService, _, _ := createTestAuthService()

	registerRes := MustCreateUserAndProfile(authService, "test1234")

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	firstName := testutils.GenerateRandomString(8)
	lastName := testutils.GenerateRandomString(8)

	profile, appErr := userService.UpdateProfileName(context.Background(), &userv1.UpdateProfileNameRequest{
		ProfileId: registerRes.Profile.Id,
		FirstName: firstName,
		LastName:  lastName,
	}, &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.Id != registerRes.Profile.Id {
		t.Fatalf("expected profile id %v, got %v", registerRes.Profile.Id, profile.Id)
	}
	if profile.FirstName != firstName {
		t.Fatalf("expected first name %v, got %v", firstName, profile.FirstName)
	}
	if profile.LastName != lastName {
		t.Fatalf("expected last name %v, got %v", lastName, profile.LastName)
	}
}

func TestUserService_UpdateProfileAvatarUrl(t *testing.T) {
	userService, _, serviceUtils := createTestUserService()
	authService, _, _ := createTestAuthService()

	registerRes := MustCreateUserAndProfile(authService, "test1234")

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	avatarURL := fmt.Sprintf("https://example.com/%s.png", testutils.GenerateRandomString(8))

	profile, appErr := userService.UpdateProfileAvatarUrl(context.Background(), &userv1.UpdateProfileAvatarUrlRequest{
		ProfileId: registerRes.Profile.Id,
		AvatarUrl: avatarURL,
	}, &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.Id != registerRes.Profile.Id {
		t.Fatalf("expected profile id %v, got %v", registerRes.Profile.Id, profile.Id)
	}
	if profile.AvatarUrl != avatarURL {
		t.Fatalf("expected avatar url %v, got %v", avatarURL, profile.AvatarUrl)
	}
}

func TestUserService_GetUser(t *testing.T) {
	userService, _, serviceUtils := createTestUserService()
	authService, _, _ := createTestAuthService()

	registerRes := MustCreateUserAndProfile(authService, "test1234")

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	getUser, appErr := userService.GetUser(context.Background(), &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if getUser == nil {
		t.Fatal("expected user, got nil")
	}
	if getUser.Email != registerRes.User.Email {
		t.Fatalf("expected email %v, got %v", registerRes.User.Email, getUser.Email)
	}
	if getUser.Id != registerRes.User.Id {
		t.Fatalf("expected user id %v, got %v", registerRes.User.Id, getUser.Id)
	}
}
