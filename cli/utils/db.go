// Package utils
package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// MustConnectDB connects to database with default config
func MustConnectDB(host string, port int, user, pass, db, sslMode string) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, pass, host, port, db, sslMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse DSN: %v", err))
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create pool: %v", err))
	}

	// Test connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		panic(fmt.Sprintf("Database unreachable: %v", err))
	}

	return pool
}

func CreateDatabase(pool *pgxpool.Pool, name string) error {
	ctx := context.Background()

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
