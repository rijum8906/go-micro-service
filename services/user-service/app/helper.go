package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/services/user/app/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (a *Application) DB() *pgxpool.Pool           { return a.infra.database }
func (a *Application) Cache() *redis.Client        { return a.infra.cache }
func (a *Application) Logger() *zap.Logger         { return a.utils.logger }
func (a *Application) Config() *config.Env         { return a.config }
func (a *Application) GRPCServer() *grpc.Server    { return a.server }
func (a *Application) BrokerCLient() broker.Client { return a.infra.brokerClient }
