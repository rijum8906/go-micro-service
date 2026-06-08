// Package session
package session

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	mock_broker "github.com/rijum8906/relay/packages/core/broker/mocks"
	"github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	"github.com/rijum8906/relay/services/user/app"
	"github.com/rijum8906/relay/services/user/app/config"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/helper"
	"go.uber.org/zap"
)

type SessionService struct {
	// Core
	DBPool          *pgxpool.Pool
	DBQ             *db.Queries
	BrokerPublisher broker.Publisher
	Helper          *helper.ServiceHelper

	// Utils
	HashService  *hash.HashService
	TokenManager token.TokenManager
	Logger       *zap.Logger

	// Config
	Config *config.Env
}

func New() (sessionv1.SessionServiceServer, *apperror.AppError) {
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

	return &SessionService{
		DBPool:          application.DB(),
		DBQ:             q,
		TokenManager:    application.TokenManager(),
		BrokerPublisher: publisher,
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 10,
		}),
		Logger: application.Logger(),
		Config: application.Config(),
		Helper: helper,
	}, nil
}

func NewForTest() *SessionService {
	// Manual configuration for testing
	config := config.Env{
		CoreEnv: coreenv.CoreEnv{
			AppEnv:          "test",
			AppName:         "user-service",
			JWTSecret:       "jwt-secret",
			ScopedSecret:    "scoped-secret",
			SessionTTL:      time.Minute,
			RefreshTokenTTL: time.Minute,
			ScopedTokenTTL:  time.Minute,
		},
	}

	// Test DB pool
	dbPool := testutils.MustConnectDB(
		testutils.WithDBName(testutils.GetTestDBName(config.AppName)),
		testutils.WithHost("localhost"),
		testutils.WithPort(5433),
	)
	q := db.New(dbPool)

	// redis clent
	redisClient := testutils.MustConnectRedis()

	// Token manager
	tokenManager := token.NewTokenManager(token.Config{
		JwtSecret:      []byte(config.JWTSecret),
		SessionTTL:     config.SessionTTL,
		ScopedTokenTTL: config.ScopedTokenTTL,
		ScopedSecret:   []byte(config.ScopedSecret),
	}, redisClient)

	// Mock publisher for publishing session events
	publisher := &mock_broker.MockPublisher{}

	// Hash service
	hashService := hash.NewHashService(hash.Config{
		BcryptCost: 10,
	})

	// Logger
	logger := corelogger.NewDevLogger()

	// Create a helper
	helper := helper.GetHelperForTest(dbPool, q, logger)

	return &SessionService{
		DBPool:          dbPool,
		DBQ:             q,
		HashService:     hashService,
		BrokerPublisher: publisher,
		TokenManager:    tokenManager,
		Logger:          logger,
		Config:          &config,
		Helper:          helper,
	}
}
