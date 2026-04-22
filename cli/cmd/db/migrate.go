package dbcmd

import (
	"fmt"

	"github.com/rijum8906/relay/cli/utils"
	"github.com/spf13/cobra"
)

var (
	Config  *utils.Environment
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
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas", "migrate", "apply",
			"--url", utils.GetDBURL(Config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas",
			"migrate", "status",
			"--url", utils.GetDBURL(Config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new atlas migration",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas", "migrate", "diff", nameVar,
			"--dir", utils.GetMigrationDir(),
			"--to", utils.GetSchemaDir(),
			"--dev-url", utils.GetDevDBURL(Config))
	},
}

var migrateRehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migrations",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas", "migrate", "rehash",
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
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas", "migrate", "down", fmt.Sprint(count),
			"--url", utils.GetDBURL(Config),
			"--dir", utils.GetMigrationDir(),
			"--dev-url", utils.GetDevDBURL(Config))
	},
}

func init() {
	if utils.IsServiceDir() {
		Config = utils.MustLoadEnv()
	}

	migrateCreateCmd.Flags().StringVarP(&nameVar, "name", "n", "", "Migration name")
	migrateRollbackCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of migrations to rollback")

	// Mark as required
	migrateCreateCmd.MarkFlagRequired("name")
	migrateRollbackCmd.MarkFlagRequired("count")

	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateRehashCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCommand)

	DBCMd.AddCommand(migrateCmd)
}
