package session

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
)

type TestSuite struct {
	*SessionService
	*auth.AuthService
	Ctx context.Context
}

func NewTestSuite() *TestSuite {
	return &TestSuite{
		Ctx:            context.Background(),
		SessionService: NewForTest(),
		AuthService:    auth.NewForTest(),
	}
}

func (suite *TestSuite) TearDownSuite() {
	suite.AuthService.DBPool.Close()      // Close the database pool
	suite.AuthService.RedisClient.Close() // Close the Redis client
	suite.SessionService.DBPool.Close()   // Close the database pool
}

func (suite *TestSuite) CreateUser(t *testing.T, email, password string) (*db.User, *db.Profile, error) {
	passwordHash, appErr := suite.SessionService.HashService.Hash(password)
	if appErr != nil {
		return nil, nil, appErr
	}
	user, err := suite.SessionService.DBQ.CreateUser(suite.Ctx, db.CreateUserParams{
		Email: email,
		PasswordHash: pgtype.Text{
			String: passwordHash,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	t.Cleanup(func() {
		suite.SessionService.DBQ.DeleteUserHard(context.Background(), user.ID)
	})

	profile, err := suite.SessionService.DBQ.CreateProfile(suite.Ctx, db.CreateProfileParams{
		UserID: user.ID,
	})
	if err != nil {
		return nil, nil, err
	}
	t.Cleanup(func() {
		suite.SessionService.DBQ.DeleteProfileHard(context.Background(), profile.ID)
	})
	return &user, &profile, nil
}

func (suite *TestSuite) SetClientInfoInContext() {

	clientInfo := dto.ClientInfo{
		DeviceID:   testutils.GenerateRandomString(32),
		IPAddress:  "127.0.0.1",
		UserAgent:  testutils.GenerateRandomString(32),
		ClientType: "web",
		APIVersion: "0.0.1",
		Locale:     "en-US",
		SDKVersion: "1.0.0",
		RequestID:  uuid.NewString(),
		TraceID:    uuid.NewString(),
	}

	suite.Ctx = testutils.SetClientInfoToIncomingContext(suite.Ctx, clientInfo)
}

func (suite *TestSuite) SetUserInfoInContext(accessToken string) {
	clams, err := suite.AuthService.TokenManager.ValidateAuthToken(suite.Ctx, accessToken)
	if err != nil {
		panic(err)
	}
	userInfo := dto.UserInfo{
		UserID:    clams.Subject,
		TokenID:   clams.ID,
		SessionID: clams.SessionID,
	}

	suite.Ctx = testutils.SetUserInfoToIncomingContext(suite.Ctx, userInfo)
}

func (suite *TestSuite) SetScopedTokenInfoInContext(subject string) {
	scopedToken := dto.ScopedToken{
		String:  testutils.GenerateRandomString(32),
		ID:      uuid.NewString(),
		Scope:   constants.TokenScope2FA,
		Subject: subject,
	}

	suite.Ctx = testutils.SetScopedTokenInfoToIncomingContext(suite.Ctx, scopedToken)
}
