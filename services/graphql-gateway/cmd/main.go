package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/resolver"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/middleware"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/utils"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	// Initialize env
	env, appErr := env.Load()
	if appErr != nil {
		panic(appErr)
	}

	// Create resolver with all services
	rootResolver := resolver.NewResolver(env)

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers:  rootResolver,
				Directives: resolver.NewDirectiveRoot(),
			},
		),
	)

	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		return utils.PresentError(ctx, err, os.Getenv("ENV") == "development")
	})

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
