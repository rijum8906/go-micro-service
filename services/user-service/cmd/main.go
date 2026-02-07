package main

import (
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/handler"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/storage"
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

	secureJWTService := jwt.NewScopedActionJWT(redisClient, jwt.Config{
		Secret:     env.JwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.JwtExpiration,
	})

	hashService := hash.NewService(10)

	middlewareService := middleware.NewMiddleware(middleware.Services{
		HashService: hashService,
		JwtService:  jwtService,
	})

	s3Storage, err := storage.NewS3Storage(ctx, env.StorageEndpoint, env.StorageAccessKey, env.StorageSecretKey, env.StorageBucket, env.StorageEndpoint)
	if err != nil {
		panic(err)
	}

	authService := services.NewAuth(db.New(pgPool), &services.UtilsConfig{
		HashService:      hashService,
		JwtService:       jwtService,
		SecureJWTService: secureJWTService,
		Storage:          s3Storage,
	}, env)
	userService := services.NewUserService(db.New(pgPool), &services.UtilsConfig{
		HashService:      hashService,
		JwtService:       jwtService,
		SecureJWTService: secureJWTService,
		Storage:          s3Storage,
	}, env)
	// server logic starts here...

	router := gin.Default()
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     env.CorsAllowedOrigins,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	apiRouterV1 := router.Group("/api/v1")

	authRouter := apiRouterV1.Group("/auth")
	userRouter := apiRouterV1.Group("/users")

	handler.RegisterHandlers(authRouter, authService)
	handler.RegisterUserHandlers(userRouter, userService, middlewareService)

	err = router.Run(fmt.Sprintf(":%s", env.AppPort))
	if err != nil {
		panic(err)
	}
}
