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
}

func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, *errors.AppError) {
	cfg.applyDefaults()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	poolConfig.MaxConns = int32(cfg.Options.MaxConns)
	poolConfig.MinConns = int32(cfg.Options.MinConns)
	poolConfig.MaxConnIdleTime = cfg.Options.MaxConnIdleTime

	if cfg.Options.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.Options.MaxConnLifetime
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.ErrDBConnection.WithInternal(err)
	}

	// Verification
	if err := pool.Ping(ctx); err != nil {
		return nil, errors.ErrDBConnection.WithInternal(err)
	}

	return pool, nil
}
