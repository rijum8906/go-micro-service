package command

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	env "github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/rijum8906/relay/packages/core/db"
)

func CreateNewDatabase(cfg *env.CoreEnv, name string) *apperror.AppError {
	ctx := context.Background()
	pool, appErr := db.Connect(ctx, db.Config{
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
		return apperror.ErrInternal.WithDetail("error", err.Error())
	}

	if exists {
		return nil // Already exists
	}

	// Create database (can't use parameters for database names)
	sql := "CREATE DATABASE " + name
	_, err = pool.Exec(ctx, sql)
	if err != nil {
		return apperror.ErrInternal.WithDetail("error", err.Error())
	}

	return nil
}

func GetDBURL(cfg *env.CoreEnv) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func GetDevDBURL(cfg *env.CoreEnv) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		"dev_"+cfg.DBName,
		cfg.DBSSLMode,
	)
}

func GetMigrationDir() string {
	return "file://db/migrations"
}

func GetSchemaDir() string {
	return "file://db/schema.sql"
}
