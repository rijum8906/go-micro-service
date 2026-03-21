// Package db provides database connection logic with PostgreSQL
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	SSLMode     string
	RetryCounts int
}

// Connect creates a connection pool and returns pgxpool.Pool
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	// 1. Parse config to set pool limits
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to parse database DSN")
	}

	// Pool configurations
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	// 2. Initialize the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to initialize database pool")
	}

	// 3. Retry Logic with Ping
	retryCounts := 5
	if cfg.RetryCounts > 0 {
		retryCounts = cfg.RetryCounts
	}

	var lastErr error
	for i := 0; i < retryCounts; i++ {
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, apperror.ErrInternal.WithMessage("Database connection cancelled by context")
		default:
			// Ping to ensure connection is actually alive
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := pool.Ping(pingCtx)
			cancel()

			if err == nil {
				return pool, nil // Success!
			}

			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	// connection failed. Close the pool and return error
	pool.Close()
	return nil, apperror.ErrInternal.WithMessage(fmt.Sprintf("Database unreachable: %v", lastErr))
}
