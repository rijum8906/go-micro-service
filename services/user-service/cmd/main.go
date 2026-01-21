package main

import (
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
)

func main() {
	env, err := env.Load()
	if err != nil {
		panic(err)
	}

	_, err = postgres.Connect(postgres.Config{
		Host:     env.DBHost,
		Port:     env.DBPort,
		User:     env.DBUser,
		Password: env.DBPassword,
		Database: env.DBName,
		SslMode:  false,
	})
	if err != nil {
		panic(err)
	}

	redis := redis.Connect(redis.Config{
		Database: env.RedisDatabase,
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		User:     env.RedisUser,
		Password: env.RedisPassword,
	})
	_ = jwt.NewService(redis, jwt.Config{
		Secret:     env.JwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.JwtExpiration,
	})
}
