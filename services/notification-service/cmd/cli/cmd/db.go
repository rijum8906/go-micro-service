package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	coreenv "github.com/rijum8906/relay/packages/core/env"
	migrations "github.com/rijum8906/relay/services/notification-service/db"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
	Run:   notImplemented,
}

var atlasCmd = &cobra.Command{
	Use:   "atlas",
	Short: "Atlas database commands",
	Run:   notImplemented,
}

var atlasApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply atlas migrations",
	Run:   notImplemented,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	Run:   notImplemented,
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new atlas migration",
	Run:   notImplemented,
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Apply schema actions",
	Run:   notImplemented,
}

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migration checksums",
	Run:   notImplemented,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback atlas migrations",
	Run:   notImplemented,
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset atlas migrations",
	Run:   notImplemented,
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "SQL database commands",
	Run:   notImplemented,
}

var sqlApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply SQL migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Applying embedded SQL migration..")
		err := applyEmbeddedMigrations(dbURL())
		if err != nil {
			return fmt.Errorf("apply embedded SQL migrations: %w", err)
		}
		fmt.Println("Embedded SQL migrations applied")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(atlasCmd)
	dbCmd.AddCommand(sqlCmd)

	atlasCmd.AddCommand(atlasApplyCmd)
	atlasCmd.AddCommand(statusCmd)
	atlasCmd.AddCommand(newCmd)
	atlasCmd.AddCommand(schemaCmd)
	atlasCmd.AddCommand(rehashCmd)
	atlasCmd.AddCommand(rollbackCmd)
	atlasCmd.AddCommand(resetCmd)

	sqlCmd.AddCommand(sqlApplyCmd)

	newCmd.Flags().String("name", "", "migration name")
	rollbackCmd.Flags().Int("count", 1, "number of migrations to roll back")
}

func dbURL() string {
	cfg := coreenv.MustLoad()

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func applyEmbeddedMigrations(databaseURL string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	allMigrations, err := migrations.All()
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	for _, migration := range allMigrations {
		statements := splitSQLStatements(migration.Content)
		for _, statement := range statements {
			if _, err := pool.Exec(ctx, statement); err != nil {
				if shouldIgnoreMigrationError(err) {
					continue
				}

				return fmt.Errorf("apply migration %s: %w", migration.Name, err)
			}
		}
	}

	return nil
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
