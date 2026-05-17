package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (a *Application) DB() *pgxpool.Pool                         { return a.infra.database }
func (a *Application) Cache() *redis.Client                      { return a.infra.cache }
func (a *Application) Logger() *zap.Logger                       { return a.utils.logger }
func (a *Application) Config() *config.Env                       { return a.config }
func (a *Application) BrokerClient() broker.Client               { return a.infra.brokerClient }
func (a *Application) TemplateManager() template.TemplateManager { return a.utils.tm }
func (a *Application) GRPCServer() *grpc.Server                  { return a.server }
func (a *Application) MailerConfig() mailer.Config               { return a.utils.mailerConfig }

// States

func (a *Application) IsLoggerLoaded() bool { return a.state.isLogggerLoaded }

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
