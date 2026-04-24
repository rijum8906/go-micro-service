// Package utils
package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DevDBName = "dev_db"

// ConnectDB connects to the database and verifies that it is reachable.
func ConnectDB(port int, user, pass, db, sslMode string) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, pass, "localhost", port, db, sslMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	// Test connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database is unreachable: %w", err)
	}

	return pool, nil
}

func CreateDatabase(pool *pgxpool.Pool, name string) error {
	ctx := context.Background()

	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		name,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check whether database %q exists: %w", name, err)
	}

	if exists {
		return nil // Already exists
	}

	// Create database (can't use parameters for database names)
	sql := "CREATE DATABASE " + name
	_, err = pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("create database %q: %w", name, err)
	}

	return nil
}
