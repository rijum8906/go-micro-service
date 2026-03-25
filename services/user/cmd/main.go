package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/cache"
	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	internaldb "github.com/rijum8906/relay/services/user/internal/db"
	handlergrpc "github.com/rijum8906/relay/services/user/internal/handlers/grpc"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	userrepo "github.com/rijum8906/relay/services/user/internal/repository/user"
	authservice "github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func bootstrap() (*env.Config, *apperror.AppError) {
	config, appErr := env.Load()
	if appErr != nil {
		return nil, appErr
	}

	return config, nil
}

func initTokenManager(env *env.Config, redisClient *redis.Client) *token.TokenManager {
	return token.NewTokenManager(env.JWTSecret, env.ScopedSecret, redisClient)
}

func initPostgres(env *env.Config, ctx context.Context) (*pgxpool.Pool, *apperror.AppError) {
	return db.Connect(ctx, db.Config{
		Host:        env.DBHost,
		Port:        env.DBPort,
		User:        env.DBUser,
		Password:    env.DBPassword,
		DBName:      env.DBName,
		SSLMode:     env.DBSSLMode,
		RetryCounts: 5,
	})
}

func initRedis(env *env.Config, ctx context.Context) (*redis.Client, *apperror.AppError) {
	return cache.Connect(ctx, cache.Config{
		Host:        env.RedisHost,
		Port:        env.RedisPort,
		DB:          0,
		Password:    env.RedisPass,
		RetryCounts: 5,
	})
}

func initGRPCServer(env *env.Config, authHandler *handlergrpc.AuthHandler) (*grpc.Server, net.Listener, *apperror.AppError) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.Port))
	if err != nil {
		return nil, nil, apperror.ErrInternal.WithMessage("failed to listen for gRPC server").WithDetail("error", err.Error())
	}

	server := grpc.NewServer()
	authv1.RegisterAuthServiceServer(server, authHandler)
	reflection.Register(server)

	return server, listener, nil
}

func main() {
	ctx := context.Background()

	env, appErr := bootstrap()
	if appErr != nil {
		log.Fatal(appErr)
	}

	postgresPool, appErr := initPostgres(env, ctx)
	if appErr != nil {
		log.Fatal(appErr)
	}
	defer postgresPool.Close()

	redisClient, appErr := initRedis(env, ctx)
	if appErr != nil {
		log.Fatal(appErr)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	tokenManager := initTokenManager(env, redisClient)
	hashService := hash.NewHashService(hash.Config{})
	dbQueries := internaldb.New(postgresPool)
	repos := utils.Repos{
		User:    userrepo.NewAuthRepository(dbQueries),
		Profile: profile.NewProfileRepository(dbQueries),
		Session: session.NewSessionRepository(dbQueries),
	}
	authHandler := handlergrpc.NewAuthHandler(
		authservice.NewAuthService(&repos, utils.NewUtils(tokenManager, hashService)),
	)

	server, listener, appErr := initGRPCServer(env, authHandler)
	if appErr != nil {
		log.Fatal(appErr)
	}

	log.Printf("gRPC server listening on %s", listener.Addr().String())
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
