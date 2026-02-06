package main

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/handler"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

func main() {
	// Initialize a global background context
	ctx := context.Background()

	env, err := env.Load()
	if err != nil {
		panic(err)
	}

	// Pass context to Postgres connection
	pgPool := postgres.Connect(ctx, postgres.Config{
		Host:     env.DBHost,
		Port:     env.DBPort,
		User:     env.DBUser,
		Password: env.DBPassword,
		Database: env.DBName,
	})

	// Pass context to Redis connection
	redisClient := redis.Connect(redis.Config{
		Database: env.RedisDatabase,
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		User:     env.RedisUser,
		Password: env.RedisPassword,
	})

	jwtService := jwt.NewService(redisClient, jwt.Config{
		Secret:     env.JwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.JwtExpiration,
	})

	hashService := hash.NewService(10)

	authService := services.NewAuth(db.New(pgPool), &services.UtilsConfig{
		HashService: hashService,
		JwtService:  jwtService,
	})
	// server logic starts here...

	router := gin.Default()
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Your React port
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	apiRouterV1 := router.Group("/api/v1")
	authRouter := apiRouterV1.Group("/auth")
	handler.RegisterHandlers(authRouter, authService)
	err = router.Run(":8906")
	if err != nil {
		panic(err)
	}
}
