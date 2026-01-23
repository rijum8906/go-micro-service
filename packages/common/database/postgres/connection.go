// Package postgres
package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type PGXConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	SSLMode  bool
}

func ConnectDB(ctx context.Context, cfg PGXConfig) *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%t", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
