// Package postgres contains functions for connecting to a postgres database
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	SSLMode  string // disable | require | verify-ca | verify-full
	Options  *Options
}

type Options struct {
	MaxConns         int
	MinConns         int
	MaxConnIdleTime  time.Duration
	MaxConnLifetime  time.Duration
	HealthCheckQuery string
	RetryAttempts    int
	RetryDelay       time.Duration
}

// DSN builds a postgres connection string
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// Set default values in a clean way
func (c *Config) applyDefaults() {
	if c.Options == nil {
		c.Options = &Options{}
	}
	if c.Options.MaxConns == 0 {
		c.Options.MaxConns = 10
	}
	if c.Options.MinConns == 0 {
		c.Options.MinConns = 2
	}
	if c.Options.MaxConnIdleTime == 0 {
		c.Options.MaxConnIdleTime = 5 * time.Minute
	}
	if c.Options.RetryAttempts == 0 {
		c.Options.RetryAttempts = 5
	}
	if c.Options.RetryDelay == 0 {
		c.Options.RetryDelay = 2 * time.Second
	}
}

func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, *errors.AppError) {
	cfg.applyDefaults()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	// Apply pool settings
	poolConfig.MaxConns = int32(cfg.Options.MaxConns)
	poolConfig.MinConns = int32(cfg.Options.MinConns)
	poolConfig.MaxConnIdleTime = cfg.Options.MaxConnIdleTime

	var pool *pgxpool.Pool
	var lastErr error

	// Retry Loop
	for i := 1; i <= cfg.Options.RetryAttempts; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			// Even if the pool is created, we must Ping to ensure Postgres is READY
			if err = pool.Ping(ctx); err == nil {
				return pool, nil // Success!
			}
		}

		lastErr = err
		fmt.Printf("⚠️ Database connection attempt %d/%d failed: %v. Retrying in %v...\n",
			i, cfg.Options.RetryAttempts, lastErr, cfg.Options.RetryDelay)

		// Wait before next attempt, but respect context cancellation
		select {
		case <-ctx.Done():
			return nil, errors.ErrDBConnection.WithInternal(ctx.Err())
		case <-time.After(cfg.Options.RetryDelay):
		}
	}

	return nil, errors.ErrDBConnection.WithInternal(fmt.Errorf("after %d attempts, last error: %w", cfg.Options.RetryAttempts, lastErr))
}
