package dbcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var (
	nameVar string
	count   int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateApplyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"up"},
	Short:   "Apply atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		pool, err := utils.ConnectDB(config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		if err != nil {
			return err
		}
		defer pool.Close()

		if err := utils.CreateDatabase(pool, config.DBName); err != nil {
			return err
		}
		if err := utils.CreateDatabase(pool, utils.DevDBName); err != nil {
			return err
		}

		return utils.RunCommand("atlas", "migrate", "apply",
			"--url", utils.GetDBURL(config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateSQLApply = &cobra.Command{
	Use:   "sql-apply",
	Short: "Apply SQL migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		content, err := os.ReadFile("db/schema.sql")
		if err != nil {
			return fmt.Errorf("read db/schema.sql: %w", err)
		}

		pool, err := utils.ConnectDB(config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		if err != nil {
			return err
		}
		defer pool.Close()
		if err = utils.CreateDatabase(pool, config.DBName); err != nil {
			return err
		}
		if err = utils.CreateDatabase(pool, utils.DevDBName); err != nil {
			return err
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			return fmt.Errorf("apply db/schema.sql: %w", err)
		}

		fmt.Println("\n✅ Applied SQL migrations")
		return nil
	},
}

var migrateSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Sync database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas",
			"schema", "apply",
			"--url", utils.GetDBURL(config),
			"--dev-url", utils.GetDevDBURL(config),
			"--file", utils.GetSchemaDir(),
			"--auto-approve")
	},
}

var migrateCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "schema", "clean",
			"--url", utils.GetDBURL(config))
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas",
			"migrate", "status",
			"--url", utils.GetDBURL(config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new atlas migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "migrate", "diff", nameVar,
			"--dir", utils.GetMigrationDir(),
			"--to", utils.GetSchemaDir(),
			"--dev-url", utils.GetDevDBURL(config))
	},
}

var migrateRehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsServiceDir() {
			return fmt.Errorf("must be run from a service directory")
		}

		return utils.RunCommand("atlas", "migrate", "rehash",
			"--dir", utils.GetMigrationDir())
	},
}

var migrateResetCommand = &cobra.Command{
	Use:   "reset",
	Short: "Reset atlas migrations",
	Run:   utils.NotImplemented,
}

var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadServiceConfig()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "migrate", "down", fmt.Sprint(count),
			"--url", utils.GetDBURL(config),
			"--dir", utils.GetMigrationDir(),
			"--dev-url", utils.GetDevDBURL(config))
	},
}

func loadServiceConfig() (*utils.Environment, error) {
	if !utils.IsServiceDir() {
		return nil, fmt.Errorf("must be run from a service directory")
	}

	return utils.LoadEnv()
}

func init() {
	migrateCreateCmd.Flags().StringVarP(&nameVar, "name", "n", "", "Migration name")
	migrateRollbackCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of migrations to rollback")

	// Mark as required
	_ = migrateCreateCmd.MarkFlagRequired("name")
	_ = migrateRollbackCmd.MarkFlagRequired("count")

	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateSchemaCmd)
	migrateCmd.AddCommand(migrateSQLApply)
	migrateCmd.AddCommand(migrateCleanCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateRehashCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCommand)

	DBCMd.AddCommand(migrateCmd)
}
