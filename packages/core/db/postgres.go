// Package db defines the database connection for the application
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

// Connect creates a connection pool
func Connect(ctx context.Context, cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Could not open database connection")
	}

	// DB connection configs
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Retry Logic
	retryCounts := 5
	if cfg.RetryCounts > 0 {
		retryCounts = cfg.RetryCounts
	}

	var lastErr error
	for i := 0; i < retryCounts; i++ {
		// Check if the parent context (e.g., system shutdown) is already done
		select {
		case <-ctx.Done():
			return nil, apperror.ErrInternal.WithMessage("Database connection cancelled by context")
		default:
			// Fresh timeout for this specific ping attempt
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := db.PingContext(pingCtx)
			cancel()

			if err == nil {
				return db, nil
			}

			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return nil, apperror.ErrInternal.WithMessage(fmt.Sprintf("Database unreachable: %v", lastErr))
}
