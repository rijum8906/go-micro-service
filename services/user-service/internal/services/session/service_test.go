package session_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type noopPublisher struct{}

func (noopPublisher) PublishJSON(string, any) *apperror.AppError {
	return nil
}

func createTestAuthService() (auth.AuthService, *utils.Repos, *utils.ServiceUtils) {
	repos := utils.NewTestRepos()
	serviceUtils := utils.NewTestServiceUtils()
	config := testutils.NewTestEnv()
	service, appErr := auth.NewAuthService(repos, serviceUtils, config, noopPublisher{})
	if appErr != nil {
		panic(appErr)
	}
	return service, repos, serviceUtils
}

func createTestSessionService() (session.SessionService, *utils.Repos, *utils.ServiceUtils) {
	repos := utils.NewTestRepos()
	serviceUtils := utils.NewTestServiceUtils()
	config := testutils.NewTestEnv()
	service, appErr := session.NewSessionService(repos, serviceUtils, config)
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

func TestSessionService_GetCurrentSession(t *testing.T) {
	sessionService, _, serviceUtils := createTestSessionService()
	authService, _, _ := createTestAuthService()

	password := "test1234"

	registerRes := MustCreateUserAndProfile(authService, password)

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	currentSession, appErr := sessionService.GetCurrentSession(context.Background(), &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if currentSession == nil {
		t.Fatal("expected session, got nil")
	}
	if currentSession.UserId != registerRes.User.Id {
		t.Fatalf("expected user id %v, got %v", registerRes.User.Id, currentSession.UserId)
	}
	if currentSession.Id != claims.ID {
		t.Fatalf("expected session id %v, got %v", claims.ID, currentSession.Id)
	}
}

func TestSessionService_RevokeSession(t *testing.T) {
	sessionService, _, serviceUtils := createTestSessionService()
	authService, _, _ := createTestAuthService()

	password := "test1234"

	registerRes := MustCreateUserAndProfile(authService, password)

	oldAccessToken := registerRes.Tokens.AccessToken.Value

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	revokeSession, appErr := sessionService.RevokeSession(context.Background(), &sessionv1.RevokeSessionRequest{
		TokenToRevoke: registerRes.Tokens.RefreshToken.Value,
	}, &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if revokeSession == nil {
		t.Fatal("expected session, got nil")
	}
	if !revokeSession.Success {
		t.Fatal("expected success, got false")
	}

	_, appErr = serviceUtils.TokenManager.ValidateAuthToken(context.Background(), oldAccessToken)
	if appErr == nil {
		t.Fatal("expected token to be revoked")
	}
}

func TestSessionService_RevokeAllSessions(t *testing.T) {
	sessionService, _, serviceUtils := createTestSessionService()
	authService, _, _ := createTestAuthService()

	password := "test1234"

	registerRes := MustCreateUserAndProfile(authService, password)

	oldAccessToken := registerRes.Tokens.AccessToken.Value

	claims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	revokeSession, appErr := sessionService.RevokeAllSessions(context.Background(), &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: claims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if revokeSession == nil {
		t.Fatal("expected session, got nil")
	}
	if !revokeSession.Success {
		t.Fatal("expected success, got false")
	}

	_, appErr = serviceUtils.TokenManager.ValidateAuthToken(context.Background(), oldAccessToken)
	if appErr == nil {
		t.Fatal("expected token to be revoked")
	}
}

func TestSessionService_RevokeOtherSessions(t *testing.T) {
	sessionService, _, serviceUtils := createTestSessionService()
	authService, _, _ := createTestAuthService()

	password := "test1234"

	registerRes := MustCreateUserAndProfile(authService, password)

	currentClaims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		registerRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	loginRes, appErr := authService.Login(context.Background(), &authv1.LoginRequest{
		Email:    registerRes.User.Email,
		Password: password,
	}, testutils.GenerateClientInfo())
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	otherClaims, appErr := serviceUtils.TokenManager.ValidateAuthToken(
		context.Background(),
		loginRes.Tokens.AccessToken.Value,
	)
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}

	revokeOtherSessions, appErr := sessionService.RevokeOtherSessions(context.Background(), &sessionv1.RevokeOtherSessionsRequest{
		CurrentSessionId: currentClaims.ID,
	}, &dto.UserInfo{
		UserID:    registerRes.User.Id,
		SessionID: currentClaims.ID,
	})
	if appErr != nil {
		t.Fatalf("expected nil, got: %v", appErr)
	}
	if revokeOtherSessions == nil {
		t.Fatal("expected session, got nil")
	}
	if !revokeOtherSessions.Success {
		t.Fatal("expected success, got false")
	}

	_, appErr = serviceUtils.TokenManager.ValidateAuthToken(context.Background(), registerRes.Tokens.AccessToken.Value)
	if appErr != nil {
		t.Fatalf("expected current token to remain valid, got: %v", appErr)
	}

	_, appErr = serviceUtils.TokenManager.ValidateAuthToken(context.Background(), loginRes.Tokens.AccessToken.Value)
	if appErr == nil {
		t.Fatal("expected other token to be revoked")
	}

	if otherClaims.ID == currentClaims.ID {
		t.Fatal("expected different session ids for current and other sessions")
	}
}
