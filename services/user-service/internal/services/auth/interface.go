// Package auth
package auth

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	mock_broker "github.com/rijum8906/relay/packages/core/broker/mocks"
	"github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app"
	"github.com/rijum8906/relay/services/user/app/config"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/helper"
	"go.uber.org/zap"
)

type AuthService struct {
	// Core
	DBPool              *pgxpool.Pool
	DBQ                 *db.Queries
	RedisClient         *redis.Client
	UserClient          userv1.UserServiceClient
	OrgOpenFGAPublisher broker.Publisher
	Helper              *helper.ServiceHelper

	// Utils
	TokenManager token.TokenManager
	HashService  *hash.HashService
	Logger       *zap.Logger

	// Config
	Config *config.Env
}

func New() (authv1.AuthServiceServer, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	q := db.New(application.DB())

	publisher := broker.NewPublisher(application.BrokerCLient().GetClient())

	helper, appErr := helper.GetHelper()
	if appErr != nil {
		return nil, appErr
	}

	return &AuthService{
		DBPool:              application.DB(),
		DBQ:                 q,
		RedisClient:         application.Cache(),
		TokenManager:        application.TokenManager(),
		OrgOpenFGAPublisher: publisher,
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 8,
		}),
		Logger: application.Logger(),
		Config: application.Config(),
		Helper: helper,
	}, nil
}

func NewForTest() *AuthService {
	config := &config.Env{
		CoreEnv: coreenv.CoreEnv{
			AppName:         "user-service",
			AppEnv:          "test",
			JWTSecret:       "jwt-secret",
			ScopedSecret:    "scoped-secret",
			SessionTTL:      time.Minute,
			RefreshTokenTTL: time.Minute,
			ScopedTokenTTL:  time.Minute,
		},
	}

	dbPool := testutils.MustConnectDB(
		testutils.WithDBName(testutils.GetTestDBName(config.AppName)),
		testutils.WithHost("localhost"),
		testutils.WithPort(5433),
	)
	q := db.New(dbPool)

	redisClient := testutils.MustConnectRedis()

	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(config.JWTSecret),
		SessionTTL:     config.SessionTTL,
		ScopedTokenTTL: config.ScopedTokenTTL,
		ScopedSecret:   []byte(config.ScopedSecret),
	}, redisClient)

	hashService := hash.NewHashService(hash.Config{
		BcryptCost: 10,
	})

	publisher := &mock_broker.MockPublisher{}

	logger := corelogger.NewDevLogger()

	helper := helper.GetHelperForTest(dbPool, q, logger)

	return &AuthService{
		DBPool:              dbPool,
		DBQ:                 q,
		RedisClient:         redisClient,
		TokenManager:        tokenManager,
		OrgOpenFGAPublisher: publisher,
		HashService:         hashService,
		Logger:              logger,
		Config:              config,
		Helper:              helper,
	}
}
