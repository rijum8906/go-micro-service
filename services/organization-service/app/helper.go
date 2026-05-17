package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/coreenv"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (a *Application) DB() *pgxpool.Pool                    { return a.infra.database }
func (a *Application) Cache() *redis.Client                 { return a.infra.cache }
func (a *Application) Logger() *zap.Logger                  { return a.utils.logger }
func (a *Application) Config() *config.Env                  { return a.config }
func (a *Application) GRPCServer() *grpc.Server             { return a.server }
func (a *Application) UserClient() userv1.UserServiceClient { return a.clients.UserClient }

// Testing

// TestLogger returns a logger for testing
func (a *Application) TestLogger() *zap.Logger {
	logger, appErr := initLogger(&config.Env{
		CoreEnv: coreenv.CoreEnv{
			AppEnv:       "test",
			LogLevel:     "DEBUG",
			LogFile:      "test",
			EnableJSON:   false,
			EnableCaller: true,
			EnableStack:  false,
		},
	})
	if appErr != nil {
		panic(appErr)
	}
	return logger
}
