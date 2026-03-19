package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/common/database/postgres"
	"github.com/rijum8906/relay/packages/common/database/redis"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/hash"
	"github.com/rijum8906/relay/packages/common/jwt"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	handlers "github.com/rijum8906/relay/services/user-service/internal/api/handlers/grpc"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	dbRoot "github.com/rijum8906/relay/services/user-service/internal/db"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/services/account"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/services/profile"
	"github.com/rijum8906/relay/services/user-service/internal/services/storage"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type infrastructure struct {
	env         *env.Env
	pgPool      *pgxpool.Pool
	redisClient *goredis.Client
}

type application struct {
	authHandler *handlers.AuthHandler
}

func run() error {
	ctx := context.Background()

	infra, err := bootstrapInfrastructure(ctx)
	if err != nil {
		return err
	}
	defer infra.pgPool.Close()
	defer infra.redisClient.Close()

	app := bootstrapApplication(ctx, infra)

	return serveGRPC(infra.env, app)
}

func bootstrapInfrastructure(ctx context.Context) (*infrastructure, error) {
	appEnv, loadErr := env.Load()
	if loadErr != nil {
		return nil, fmt.Errorf("load environment: %w", toError(loadErr))
	}

	pgPool, err := connectPostgres(ctx, appEnv)
	if err != nil {
		return nil, err
	}

	if err := dbRoot.RunMigrations(ctx, pgPool); err != nil {
		pgPool.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	redisClient, err := connectRedis(ctx, appEnv)
	if err != nil {
		pgPool.Close()
		return nil, err
	}

	return &infrastructure{
		env:         appEnv,
		pgPool:      pgPool,
		redisClient: redisClient,
	}, nil
}

func connectPostgres(ctx context.Context, appEnv *env.Env) (*pgxpool.Pool, error) {
	pgPool, err := postgres.Connect(ctx, postgres.Config{
		Host:     appEnv.DBHost,
		Port:     appEnv.DBPort,
		User:     appEnv.DBUser,
		Password: appEnv.DBPassword,
		Database: appEnv.DBName,
		SSLMode:  "disable",
		Options: &postgres.Options{
			RetryAttempts: 5,
			RetryDelay:    2 * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", toError(err))
	}

	return pgPool, nil
}

func connectRedis(ctx context.Context, appEnv *env.Env) (*goredis.Client, error) {
	redisClient, err := redis.Connect(ctx, redis.Config{
		Database: appEnv.RedisDatabase,
		Host:     appEnv.RedisHost,
		Port:     appEnv.RedisPort,
		User:     appEnv.RedisUser,
		Password: appEnv.RedisPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", toError(err))
	}

	return redisClient, nil
}

func bootstrapApplication(ctx context.Context, infra *infrastructure) *application {
	hashService := hash.NewService(infra.env.BcryptCost)
	jwtService, scopedJWTService := buildJWTServices(infra.env, infra.redisClient)
	middlewareService := middleware.NewMiddleware(middleware.Services{
		HashService: hashService,
		JwtService:  jwtService,
	})

	s3Storage := storage.NewS3StorageService(
		ctx,
		infra.env.StorageEndpoint,
		infra.env.StorageAccessKey,
		infra.env.StorageSecretKey,
		infra.env.StorageBucket,
		infra.env.StoragePublicKey,
	)
	s3Storage.CreateBucket(ctx, infra.env.StorageBucket)

	utilsCfg := &utils.UtilsConfig{
		HashService:      hashService,
		JwtService:       jwtService,
		SecureJWTService: scopedJWTService,
		Storage:          s3Storage,
	}

	queries := db.New(infra.pgPool)
	authService := auth.NewAuthService(queries, utilsCfg, infra.env)
	accountService := account.NewAccountService(queries, utilsCfg, infra.env)
	profileService := profile.NewProfileService(queries, utilsCfg, infra.env)

	authHandler := handlers.NewAuthHandler(&handlers.Services{
		AuthService:    authService,
		AccountService: accountService,
		Profileservice: profileService,
	}, middlewareService)

	return &application{
		authHandler: authHandler,
	}
}

func buildJWTServices(appEnv *env.Env, redisClient *goredis.Client) (jwt.Service, jwt.ScopedActionJWT) {
	jwtService := jwt.NewService(redisClient, jwt.Config{
		Secret:     appEnv.JwtSecret,
		Issuer:     appEnv.JwtIssuer,
		Expiration: appEnv.JwtExpiration,
	})

	scopedJWTService := jwt.NewScopedActionJWT(redisClient, jwt.Config{
		Secret:     appEnv.ScopedJwtSecret,
		Issuer:     appEnv.JwtIssuer,
		Expiration: appEnv.ScopedJwtExpiration,
	})

	return jwtService, scopedJWTService
}

func serveGRPC(appEnv *env.Env, app *application) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", appEnv.AppPort))
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", appEnv.AppPort, err)
	}

	server := grpc.NewServer()
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	authv1.RegisterAuthServiceServer(server, app.authHandler)
	reflection.Register(server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}

func toError(err interface{ Error() string }) error {
	var target error
	if errors.As(err, &target) {
		return target
	}

	return errors.New(err.Error())
}
