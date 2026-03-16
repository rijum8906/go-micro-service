// Package migrate contains migration commands for databases
package migrate

import (
	"context"
	"time"

	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/db"
)

func main() {
	// Initialize a global background context
	ctx := context.Background()

	env, appErr := env.Load()
	if appErr != nil {
		panic(appErr)
	}

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
	pgPool, appErr := postgres.Connect(ctx, postgresCfg)
	if appErr != nil {
		panic(appErr.Message)
	}

	err := db.RunMigrations(ctx, pgPool)
	if err != nil {
		panic(err)
	}
}
