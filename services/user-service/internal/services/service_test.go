package services_test

import (
	"context"
	"testing"

	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	env, err := env.Load()
	if err != nil {
		panic(err)
	}

	// Pass context to Postgres connection
	_ = postgres.Connect(ctx, postgres.Config{
		Host:     env.DBHost,
		Port:     env.DBPort,
		User:     env.DBUser,
		Password: env.DBPassword,
		Database: env.DBName,
	})

	// Pass context to Redis connection
	rdb := redis.Connect(redis.Config{
		Database: env.RedisDatabase,
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		User:     env.RedisUser,
		Password: env.RedisPassword,
	})

	_ = jwt.NewService(rdb, jwt.Config{
		Secret:     env.JwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.JwtExpiration,
	})
}
