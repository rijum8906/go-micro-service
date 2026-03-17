package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rijum8906/go-micro-service/packages/common/database/redis"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	"github.com/rijum8906/go-micro-service/services/graphql-gateway/graph/generated"
	"github.com/rijum8906/go-micro-service/services/graphql-gateway/graph/resolver"
	"github.com/rijum8906/go-micro-service/services/graphql-gateway/internal/client"
	"github.com/rijum8906/go-micro-service/services/graphql-gateway/internal/middleware"
	"github.com/rs/cors"
)

const (
	UserServiceBaseURl = "http://user-service:8906/api/v1"
)

func main() {
	ctx := context.Background()

	// Initialize env
	env, appErr := env.Load()
	if appErr != nil {
		panic(appErr)
	}

	logger := slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, nil), slog.LevelDebug)
	httpClient := client.NewClient(&client.ClientConfig{
		BaseURL:       UserServiceBaseURl,
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryWaitTime: 100 * time.Millisecond,
		Logger:        logger,
	})

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

	// Create resolver with all services
	rootResolver := resolver.NewResolver(
		env,
		&resolver.Services{
			JWTService: jwtService,
		},
		&resolver.HTTPClinets{
			UserServiceClient: httpClient,
		},
	)

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers:  rootResolver,
				Directives: resolver.NewDirectiveRoot(),
			},
		),
	)

	// Create Service
	middleware := middleware.NewService(env)

	// Create router
	mux := http.NewServeMux()

	// GraphQL endpoint with auth middleware
	mux.Handle("/query", middleware.AuthMiddleware(srv))

	// Playground for development
	if os.Getenv("ENV") != "production" {
		mux.Handle("/", http.RedirectHandler("/playground", http.StatusTemporaryRedirect))
		mux.Handle("/playground", playground.Handler("GraphQL Playground", "/query"))
	}

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// CORS configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   env.CorsAllowedOrigins,
		AllowedMethods:   env.CorsAllowedMethods,
		AllowedHeaders:   env.CorsAllowedHeaders,
		AllowCredentials: true,
		Debug:            os.Getenv("ENV") == "development",
	})

	// Create server
	srvHTTP := &http.Server{
		Addr:         ":" + env.AppPort,
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("GraphQL server running on http://localhost:%s/query", env.AppPort)
		log.Printf("Playground available at http://localhost:%s/playground", env.AppPort)
		if err := srvHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srvHTTP.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
