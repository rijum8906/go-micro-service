package cmd

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/env"
)

func getDBURL(cfg *env.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func getDevDBURL(cfg *env.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		"dev_"+cfg.DBName,
		cfg.DBSSLMode,
	)
}

func getMigrationDir() string {
	return "file://db/migrations"
}

func getSchemaDir() string {
	return "file://db/schema.sql"
}

func createNewDatabase(cfg *env.Config, name string) error {
	ctx := context.Background()
	pool, appErr := db.Connect(context.Background(), db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if appErr != nil {
		return appErr
	}

	// Check if database exists
	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		name,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already exists
	}

	// Create database (can't use parameters for database names)
	sql := "CREATE DATABASE " + name
	_, err = pool.Exec(ctx, sql)
	return err
}
