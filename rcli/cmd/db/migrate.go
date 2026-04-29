package dbcmd

import (
	"fmt"

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
		config, err := utils.LoadEnv()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return utils.RunCommand("atlas", "migrate", "apply",
			"--url", utils.GetDynamicDBURL(useTestDB, config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Sync database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas",
			"schema", "apply",
			"--url", utils.GetDynamicDBURL(useTestDB, config),
			"--dev-url", utils.GetDynamicDevDBURL(useTestDB, config),
			"--file", utils.GetSchemaDir(),
			"--auto-approve")
	},
}

var migrateCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "schema", "clean",
			"--url", utils.GetDynamicDBURL(useTestDB, config))
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas",
			"migrate", "status",
			"--url", utils.GetDynamicDBURL(useTestDB, config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new atlas migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "migrate", "diff", nameVar,
			"--dir", utils.GetMigrationDir(),
			"--to", utils.GetSchemaDir(),
			"--dev-url", utils.GetDynamicDevDBURL(useTestDB, config))
	},
}

var migrateRehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsServiceDir() {
			return fmt.Errorf("must be run from a service directory")
		}

		return utils.RunCommand("atlas", "migrate", "hash",
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
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		return utils.RunCommand("atlas", "migrate", "down", fmt.Sprint(count),
			"--url", utils.GetDynamicDBURL(useTestDB, config),
			"--dir", utils.GetMigrationDir(),
			"--dev-url", utils.GetDynamicDevDBURL(useTestDB, config))
	},
}

func init() {
	migrateCreateCmd.Flags().StringVarP(&nameVar, "name", "n", "", "Migration name")
	migrateRollbackCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of migrations to rollback")

	// Mark as required
	_ = migrateCreateCmd.MarkFlagRequired("name")
	_ = migrateRollbackCmd.MarkFlagRequired("count")

	// Add Test Flags
	migrateCmd.PersistentFlags().BoolVarP(&useTestDB, "test", "t", false, "Run migrations for test environment database")

	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateSchemaCmd)
	migrateCmd.AddCommand(migrateCleanCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateRehashCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCommand)

	DBCmd.AddCommand(migrateCmd)
}
