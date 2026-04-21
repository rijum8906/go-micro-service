// Package utils
package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MustConnectDB connects to database with default config
func MustConnectDB() *pgxpool.Pool {
	dsn := "postgres://test_user:test_password@localhost:5433/test_db?sslmode=disable"

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
