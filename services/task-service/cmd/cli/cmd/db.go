package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rijum8906/relay/packages/core/command"
	"github.com/rijum8906/relay/packages/core/coreenv"
	"github.com/spf13/cobra"
)

var config *coreenv.CoreEnv

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the task service database schema and migrations",
	Long:  "Manage Atlas-based migrations for the task service database.",
}

var dbAtlasCmd = &cobra.Command{
	Use:   "atlas",
	Short: "Manage Atlas-based migrations",
	Long:  "Run Atlas migration and schema operations against the configured database.",
}

var atlasDBInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize atlas database",
	Run: func(cmd *cobra.Command, args []string) {
		defaultCfg := localDBConfig("postgres")

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

func newDBAtlasApplyCmd() *cobra.Command {
	return &cobra.Command{
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
}

func newDBAtlasStatusCmd() *cobra.Command {
	return &cobra.Command{
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
}

func newDBAtlasNewCmd() *cobra.Command {
	var migrationName string

	cmd := &cobra.Command{
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

	cmd.Flags().StringVar(&migrationName, "name", "", "migration name")
	return cmd
}

func newDBAtlasSchemaCmd() *cobra.Command {
	return &cobra.Command{
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
}

func newDBAtlasRehashCmd() *cobra.Command {
	return &cobra.Command{
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
}

func newDBAtlasRollbackCmd() *cobra.Command {
	var rollbackCount int

	cmd := &cobra.Command{
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

	cmd.Flags().IntVar(&rollbackCount, "count", 1, "number of migrations to roll back")
	return cmd
}

func newDBAtlasResetCmd() *cobra.Command {
	return &cobra.Command{
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
}

func init() {
	envConfig, appErr := coreenv.Load()
	if appErr != nil {
		panic("failed to load env")
	}

	config = envConfig

	dbAtlasCmd.AddCommand(atlasDBInitCmd)
	dbAtlasCmd.AddCommand(newDBAtlasApplyCmd())
	dbAtlasCmd.AddCommand(newDBAtlasStatusCmd())
	dbAtlasCmd.AddCommand(newDBAtlasNewCmd())
	dbAtlasCmd.AddCommand(newDBAtlasSchemaCmd())
	dbAtlasCmd.AddCommand(newDBAtlasRehashCmd())
	dbAtlasCmd.AddCommand(newDBAtlasRollbackCmd())
	dbAtlasCmd.AddCommand(newDBAtlasResetCmd())

	dbCmd.AddCommand(dbAtlasCmd)

	// Keep Atlas commands at the top level as compatibility aliases for the Makefile.
	dbCmd.AddCommand(newDBAtlasApplyCmd())
	dbCmd.AddCommand(newDBAtlasStatusCmd())
	dbCmd.AddCommand(newDBAtlasNewCmd())
	dbCmd.AddCommand(newDBAtlasSchemaCmd())
	dbCmd.AddCommand(newDBAtlasRehashCmd())
	dbCmd.AddCommand(newDBAtlasRollbackCmd())
	dbCmd.AddCommand(newDBAtlasResetCmd())

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
	return namedDBURL(config.DBName)
}

func namedDBURL(name string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		config.DBUser,
		config.DBPassword,
		"localhost",
		config.DBPort,
		name,
		config.DBSSLMode,
	)
}

func migrationDir() string {
	return "file://db/migrations"
}

func schemaURL() string {
	return "file://db/schema.sql"
}

func devDBURL() string {
	return namedDBURL("dev_" + config.DBName)
}

func localDBConfig(dbName string) coreenv.CoreEnv {
	cfg := *config
	cfg.DBHost = "localhost"
	cfg.DBName = dbName
	return cfg
}

func localTestAdminConfig() coreenv.CoreEnv {
	cfg := *config
	cfg.DBHost = "localhost"
	cfg.DBPort = 5433
	cfg.DBUser = "test_user"
	cfg.DBPassword = "test_password"
	cfg.DBName = "test_db"
	cfg.DBSSLMode = "disable"
	return cfg
}

func ensureTestDatabase() error {
	testCfg := localTestAdminConfig()
	if err := command.CreateNewDatabase(&testCfg, testDBName()); err != nil {
		return fmt.Errorf("create test database %s: %w", testDBName(), err)
	}

	return nil
}
