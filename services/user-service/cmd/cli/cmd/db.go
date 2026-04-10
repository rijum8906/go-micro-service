package cmd

import (
	"fmt"
	"os"
	"os/exec"

	coreenv "github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/spf13/cobra"
)

var (
	rollbackCount int
	migrationName string
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the user service database schema and migrations",
	Long:  "Manage Atlas migrations and schema operations for the user service database.",
}

var dbApplyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"migrate"},
	Short:   "Apply pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Applying migrations...")
		if err := runCommand(
			"atlas", "migrate", "apply",
			"--url", dbURL(),
			"--dir", migrationDir(),
		); err != nil {
			return fmt.Errorf("apply migrations: %w", err)
		}

		fmt.Println("Migrations applied.")
		return nil
	},
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking migration status...")
		if err := runCommand(
			"atlas", "migrate", "status",
			"--url", dbURL(),
			"--dir", migrationDir(),
		); err != nil {
			return fmt.Errorf("check migration status: %w", err)
		}

		return nil
	},
}

var dbNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate a new migration from schema changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrationName == "" {
			return fmt.Errorf("migration name is required")
		}

		fmt.Printf("Generating migration: %s\n", migrationName)
		if err := runCommand(
			"atlas", "migrate", "diff", migrationName,
			"--dir", migrationDir(),
			"--to", schemaURL(),
			"--dev-url", devDBURL(),
		); err != nil {
			return fmt.Errorf("generate migration: %w", err)
		}

		fmt.Println("Migration generated.")
		return nil
	},
}

var dbSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Apply schema directly",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Applying schema...")
		if err := runCommand(
			"atlas", "schema", "apply",
			"--url", dbURL(),
			"--to", schemaURL(),
			"--dev-url", devDBURL(),
			"--auto-approve",
		); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}

		fmt.Println("Schema applied.")
		return nil
	},
}

var dbRehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Recalculate atlas.sum",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Rehashing migration directory...")
		if err := runCommand("atlas", "migrate", "hash", "--dir", migrationDir()); err != nil {
			return fmt.Errorf("rehash migrations: %w", err)
		}

		fmt.Println("atlas.sum updated.")
		return nil
	},
}

var dbRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Roll back applied migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rollbackCount < 1 {
			return fmt.Errorf("rollback count must be at least 1")
		}

		fmt.Printf("Rolling back %d migration(s)...\n", rollbackCount)
		if err := runCommand(
			"atlas", "migrate", "down", fmt.Sprint(rollbackCount),
			"--url", dbURL(),
			"--dir", migrationDir(),
			"--dev-url", devDBURL(),
		); err != nil {
			return fmt.Errorf("rollback migrations: %w", err)
		}

		fmt.Println("Rollback complete.")
		return nil
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the database and reapply migrations",
	Long:  "Drop all objects from the configured database and reapply the current migration set.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Cleaning database...")
		if err := runCommand(
			"atlas", "schema", "clean",
			"--url", dbURL(),
			"--auto-approve",
		); err != nil {
			return fmt.Errorf("clean database: %w", err)
		}

		fmt.Println("Reapplying migrations...")
		if err := runCommand(
			"atlas", "migrate", "apply",
			"--url", dbURL(),
			"--dir", migrationDir(),
		); err != nil {
			return fmt.Errorf("reapply migrations: %w", err)
		}

		fmt.Println("Database reset complete.")
		return nil
	},
}

func init() {
	dbNewCmd.Flags().StringVar(&migrationName, "name", "", "migration name")
	dbRollbackCmd.Flags().IntVar(&rollbackCount, "count", 1, "number of migrations to roll back")

	dbCmd.AddCommand(dbApplyCmd)
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbNewCmd)
	dbCmd.AddCommand(dbSchemaCmd)
	dbCmd.AddCommand(dbRehashCmd)
	dbCmd.AddCommand(dbRollbackCmd)
	dbCmd.AddCommand(dbResetCmd)

	rootCmd.AddCommand(dbCmd)
}

func runCommand(name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}

func dbURL() string {
	cfg := coreenv.MustLoad()

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		cfg.DBUser,
		cfg.DBPassword,
		"localhost",
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func migrationDir() string {
	return "file://db/migrations"
}

func schemaURL() string {
	return "file://db/schema.sql"
}

func devDBURL() string {
	return testutils.DevDBURL
}
