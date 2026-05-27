package app

import (
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/graphql-gateway/app/config"
	"go.uber.org/zap"
)

func (a *Application) RedisClient() *redis.Client        { return a.infra.redisClient }
func (a *Application) TokenManager() *token.TokenManager { return a.utils.token }
func (a *Application) Logger() *zap.Logger               { return a.utils.logger }
func (a *Application) Config() *config.Env               { return a.config }
