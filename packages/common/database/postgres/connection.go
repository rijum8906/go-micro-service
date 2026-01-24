package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	SSLMode  string // disable | require | verify-ca | verify-full
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

// Connect creates and returns a pgx connection pool
func Connect(ctx context.Context, cfg Config) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to parse db config: %v", err)
	}

	// Sensible defaults (you can tune later)
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Verify connection early (fail fast)
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	return pool
}
