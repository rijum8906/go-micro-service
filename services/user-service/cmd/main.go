package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	dbRoot "github.com/rijum8906/go-micro-service/services/user-service/internal/db"
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

	router := gin.Default()

	postgresCfg := postgres.Config{
		Host:     env.DBHost,
		Port:     env.DBPort,
		User:     env.DBUser,
		Password: env.DBPassword,
		Database: env.DBName,
		SSLMode:  "disable",
		Options: &postgres.Options{
			RetryAttempts: 5,
			RetryDelay:    time.Second * 2,
		},
	}
	// Pass context to Postgres connection
	pgPool, err := postgres.Connect(ctx, postgresCfg)
	if err != nil {
		panic(err.Message)
	}

	errr := dbRoot.RunMigrations(ctx, pgPool)
	if errr != nil {
		panic(errr)
	}

	redisCfg := redis.Config{
		Database: env.RedisDatabase,
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		User:     env.RedisUser,
		Password: env.RedisPassword,
	}
	// Pass context to Redis connection
	redisClient, err := redis.Connect(ctx, redisCfg)
	if err != nil {
		panic(err.Message)
	}

	jwtService := jwt.NewService(redisClient, jwt.Config{
		Secret:     env.JwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.JwtExpiration,
	})

	scopedJWTService := jwt.NewScopedActionJWT(redisClient, jwt.Config{
		Secret:     env.ScopedJwtSecret,
		Issuer:     env.JwtIssuer,
		Expiration: env.ScopedJwtExpiration,
	})

	hashService := hash.NewService(env.BcryptCost)

	middlewareService := middleware.NewMiddleware(middleware.Services{
		HashService: hashService,
		JwtService:  jwtService,
	})

	s3Storage := storage.NewS3StorageService(
		ctx,
		env.StorageEndpoint,
		env.StorageAccessKey,
		env.StorageSecretKey,
		env.StorageBucket,
		env.StoragePublicKey,
	)
	s3Storage.CreateBucket(ctx, env.StorageBucket)

	utilsCfg := &services.UtilsConfig{
		HashService:      hashService,
		JwtService:       jwtService,
		SecureJWTService: scopedJWTService,
		Storage:          s3Storage,
	}

	authService := services.NewAuth(db.New(pgPool), utilsCfg, env)
	accountService := services.NewAccountService(db.New(pgPool), utilsCfg, env)
	profileService := services.NewProfileService(db.New(pgPool), utilsCfg, env)
	// server logic starts here...

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     env.CorsAllowedOrigins,
		AllowMethods:     env.CorsAllowedMethods,
		AllowHeaders:     env.CorsAllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	apiRouterV1 := router.Group("/api/v1")

	authRouter := apiRouterV1.Group("/auth")
	profilesRouter := apiRouterV1.Group("/profiles")
	accountRouter := apiRouterV1.Group("/accounts")

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	handler.RegisterHandlers(authRouter, authService)
	handler.SetupProfilesHandlers(profilesRouter, profileService, middlewareService)
	handler.SetupAccountsHandlers(accountRouter, accountService, middlewareService)

	error := router.Run(fmt.Sprintf(":%s", env.AppPort))
	if error != nil {
		panic(error)
	}
}
