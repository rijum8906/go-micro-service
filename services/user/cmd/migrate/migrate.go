package main

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rijum8906/relay/packages/core/apperror"
	coredb "github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/env"
	migrations "github.com/rijum8906/relay/services/user/db/migrations"
)

func main() {
	ctx := context.Background()
	cfg, appErr := env.Load()
	if appErr != nil {
		panic(appErr)
	}

	pool, appErr := coredb.Connect(ctx, coredb.Config{
		Host:        cfg.DBHost,
		Port:        cfg.DBPort,
		User:        cfg.DBUser,
		Password:    cfg.DBPassword,
		DBName:      cfg.DBName,
		SSLMode:     cfg.DBSSLMode,
		RetryCounts: 5,
	})
	if appErr != nil {
		panic(appErr)
	}
	defer pool.Close()

	allMigrations, err := migrations.All()
	if err != nil {
		panic(apperror.ErrInternal.WithMessage("failed to load migrations").WithDetail("error", err.Error()))
	}

	for _, migration := range allMigrations {
		statements := splitSQLStatements(migration.Content)
		for _, statement := range statements {
			if _, err := pool.Exec(ctx, statement); err != nil {
				if shouldIgnoreMigrationError(err) {
					continue
				}
				panic(apperror.ErrInternal.WithMessage("failed to apply migration").WithDetail("migration", migration.Name).WithDetail("error", err.Error()))
			}
		}
	}
}

func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement == "" {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func shouldIgnoreMigrationError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "42P07", "42710":
		return true
	default:
		return false
	}
}
