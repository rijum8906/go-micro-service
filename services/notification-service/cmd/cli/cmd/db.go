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
	"github.com/rijum8906/relay/packages/core/command"
)

var config *env.Config

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
	Run:   command.NotImplemented,
}

var atlasCmd = &cobra.Command{
	Use:   "atlas",
	Short: "Atlas database commands",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var atlasDBInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize atlas database",
	Run: func(cmd *cobra.Command, args []string) {
		defaultCfg := *config
		defaultCfg.DBName = "postgres"

		fmt.Println("Initializing database...")

		err := command.CreateNewDatabase(&defaultCfg, config.DBName)
		if err != nil {
			fmt.Printf("failed to create database %s: %v", config.DBName, err)
		}

		err = command.CreateNewDatabase(&defaultCfg, "dev_"+config.DBName)
		if err != nil {
			fmt.Printf("failed to create database %s: %v", "dev_"+config.DBName, err)
		}

		fmt.Println("Database initialized.")
	},
}

var atlasApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply atlas migrations",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas",
			"migrate", "apply",
			"--url", command.GetDBURL(config),
			"--dir", command.GetMigrationDir())
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas",
			"migrate", "status",
			"--url", command.GetDBURL(config),
			"--dir", command.GetMigrationDir())
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new atlas migration",
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		command.RunCommand("atlas", "migrate", "diff", migrationName,
			"--dir", command.GetMigrationDir(),
			"--to", command.GetSchemaDir(),
			"--dev-url", command.GetDevDBURL(config))
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Apply schema actions",
	Run: func(cmd *cobra.Command, args []string) {
		err := command.CreateNewDatabase(config, "dev_"+config.DBName)
		if err != nil {
			panic(err)
		}
		command.RunCommand("atlas",
			"schema", "apply",
			"--url", command.GetDBURL(config),
			"--dev-url", command.GetDevDBURL(config),
			"--file", command.GetSchemaDir(),
			"--auto-approve")
	},
}

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migration checksums",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas", "migrate", "hash",
			"--dir", command.GetMigrationDir())
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback atlas migrations",
	Run: func(cmd *cobra.Command, args []string) {
		count := args[0]
		command.RunCommand("atlas", "migrate", "down", count,
			"--url", command.GetDBURL(config),
			"--dir", command.GetMigrationDir(),
			"--dev-url", command.GetDevDBURL(config))
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset atlas migrations",
	Run:   command.NotImplemented,
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "SQL database commands",
	Run:   command.NotImplemented,
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
	envConfig, appErr := env.Load()
	if appErr != nil {
		panic("failed to load env")
	}

	config = envConfig

	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(atlasCmd)
	dbCmd.AddCommand(sqlCmd)

	atlasCmd.AddCommand(atlasDBInitCmd)
	atlasCmd.AddCommand(atlasApplyCmd)
	atlasCmd.AddCommand(statusCmd)
	atlasCmd.AddCommand(newCmd)
	atlasCmd.AddCommand(schemaCmd)
	atlasCmd.AddCommand(rehashCmd)
	atlasCmd.AddCommand(rollbackCmd)
	atlasCmd.AddCommand(resetCmd)

	sqlCmd.AddCommand(sqlApplyCmd)
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
