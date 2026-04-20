package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type TestCase struct {
	name string
	req  *authv1.LoginRequest
}

func createTestAuthService() (auth.AuthService, *utils.Repos, *utils.ServiceUtils) {
	repos := utils.NewTestRepos()
	serviceUtils := utils.NewTestServiceUtils()
	service, appErr := auth.NewAuthService(repos, serviceUtils, nil, nil)
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

func Test_authService_RegisterLoginAndLogout(t *testing.T) {
	service, _, serviceUtils := createTestAuthService()

	password := "test1234"

	// Register
	authRes := MustCreateUserAndProfile(service, password)

	// Validating Generated Token
	_, appErr := serviceUtils.TokenManager.ValidateAuthToken(context.Background(),
		authRes.Tokens.AccessToken.Value)
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	// Login
	loginRes, appErr := service.Login(context.Background(), &authv1.LoginRequest{
		Email:    authRes.User.Email,
		Password: password,
	}, testutils.GenerateClientInfo())
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	// Validating Generated Token
	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(context.Background(),
		loginRes.Tokens.AccessToken.Value)
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	// Logout
	success, appErr := service.Logout(context.Background(), &dto.UserInfo{
		UserID:    loginRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}
	if !success {
		t.Errorf("expected true, got: %v", success)
	}
}

func Test_authService_RefreshToken(t *testing.T) {
	service, _, serviceUtils := createTestAuthService()

	password := "test1234"

	// Register
	registerRes := MustCreateUserAndProfile(service, password)

	// Validating Generated Token
	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(context.Background(),
		registerRes.Tokens.AccessToken.Value)
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	tokenRes, appErr := service.RefreshToken(context.Background(), &dto.UserInfo{
		UserID:      registerRes.User.Id,
		AccessToken: registerRes.Tokens.AccessToken.Value,
		SessionID:   claims.ID,
	})
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	// Validating Refresh Token
	claims2, appErr := serviceUtils.TokenManager.ValidateAuthToken(context.Background(),
		tokenRes.AccessToken.Value)
	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}

	if claims2.ID != claims.ID {
		t.Errorf("expected %v, got: %v", claims.ID, claims2.ID)
	}

	if claims2.ExpiresAt.Before(claims.ExpiresAt.Time) {
		t.Errorf("expected %v, got: %v", claims.ExpiresAt.Time, claims2.ExpiresAt.Time)
	}

	if appErr != nil {
		t.Errorf("expected nil, got: %v", appErr)
	}
}
