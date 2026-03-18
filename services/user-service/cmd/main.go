package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/packages/common/database/postgres"
	"github.com/rijum8906/relay/packages/common/database/redis"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/hash"
	"github.com/rijum8906/relay/packages/common/jwt"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	grpc_handler "github.com/rijum8906/relay/services/user-service/internal/api/handlers/grpc"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	dbRoot "github.com/rijum8906/relay/services/user-service/internal/db"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/services/storage"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize a global background context
	ctx := context.Background()

	env, err := env.Load()
	if err != nil {
		panic(err)
	}

	router := gin.New()

	// Global middlewares
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())

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

	utilsCfg := &utils.UtilsConfig{
		HashService:      hashService,
		JwtService:       jwtService,
		SecureJWTService: scopedJWTService,
		Storage:          s3Storage,
	}

	authService := auth.NewAuth(db.New(pgPool), utilsCfg, env)
	// accountService := account.NewAccountService(db.New(pgPool), utilsCfg, env)
	// profileService := profile.NewProfileService(db.New(pgPool), utilsCfg, env)
	// server logic starts here...

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	lis, errr := net.Listen("tcp", fmt.Sprintf(":%s", env.AppPort))
	if errr != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	authHandler := grpc_handler.NewAuthHandler(authService, middlewareService)
	user_servicev1.RegisterAuthServiceServer(s, authHandler)

	// Enable reflection (allows tools like Postman/grpcurl to "see" API)
	reflection.Register(s)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
