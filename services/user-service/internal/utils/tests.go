package utils

import (
	"context"
	"time"

	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

func NewTestRepos() *Repos {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("user-service")))
	querier := db.New(pool)

	return &Repos{
		User:    user.NewAuthRepository(querier),
		Profile: profile.NewProfileRepository(querier),
		Session: session.NewSessionRepository(querier),
	}
}

func NewTestServiceUtils() *ServiceUtils {
	redisClient := testutils.MustConnectRedis()
	return &ServiceUtils{
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 10,
		}),
		TokenManager: token.NewTokenManager(token.Config{
			JwtSecret:      []byte("jwt_secret"),
			SessionTTL:     time.Minute,
			ScopedSecret:   []byte("scoped_secret"),
			ScopedTokenTTL: time.Minute,
		}, redisClient),
	}
}

func MustCreateUser(repos *Repos) *db.User {
	u, appErr := repos.User.CreateUser(context.Background(), &authv1.RegisterRequest{
		Email:     testutils.GenerateRandomEmail(),
		Password:  testutils.GenerateRandomString(10),
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
	})
	if appErr != nil {
		panic(appErr)
	}

	return u
}
