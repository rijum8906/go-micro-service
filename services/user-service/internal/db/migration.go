package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. Convert pgxpool to standard *sql.DB
	// This doesn't create a new connection, it just wraps the pool
	db := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer db.Close()

	// 2. Tell goose where to find the embedded files
	goose.SetBaseFS(embedMigrations)

	// 3. Set the dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// 4. Run 'Up' migrations
	// "migrations" is the folder name inside our embedded FS
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return nil
}
