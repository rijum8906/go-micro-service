package dbcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var (
	config  *utils.Environment
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
		pool := utils.MustConnectDB(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		if err := utils.CreateDatabase(pool, config.DBName); err != nil {
			fmt.Println(err)
			return
		}
		if err := utils.CreateDatabase(pool, "dev_"+config.DBName); err != nil {
			fmt.Println(err)
			return
		}

		utils.RunCommand("atlas", "migrate", "apply",
			"--url", utils.GetDBURL(config),
			"--dir", utils.GetMigrationDir())
	},
}

var migrateSQLApply = &cobra.Command{
	Use:   "sql-apply",
	Short: "Apply SQL migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsServiceDir() {
			fmt.Println("Not in a service directory")
			return
		}
		content, err := os.ReadFile("db/schema.sql")
		if err != nil {
			fmt.Println(err)
			return
		}

		pool := utils.MustConnectDB(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		if err = utils.CreateDatabase(pool, config.DBName); err != nil {
			fmt.Println(err)
			return
		}
		if err = utils.CreateDatabase(pool, "dev_"+config.DBName); err != nil {
			fmt.Println(err)
			return
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("✅ Applied SQL migrations")
	},
}

var migrateSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Sync database schema",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas",
			"schema", "apply",
			"--url", utils.GetDBURL(config),
			"--dev-url", utils.GetDevDBURL(config),
			"--file", utils.GetSchemaDir(),
			"--auto-approve")
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("atlas",
			"migrate", "status",
			"--url", utils.GetDBURL(config),
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
			"--dev-url", utils.GetDevDBURL(config))
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
			"--url", utils.GetDBURL(config),
			"--dir", utils.GetMigrationDir(),
			"--dev-url", utils.GetDevDBURL(config))
	},
}

func init() {
	if utils.IsServiceDir() {
		config = utils.MustLoadEnv()
	}

	migrateCreateCmd.Flags().StringVarP(&nameVar, "name", "n", "", "Migration name")
	migrateRollbackCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of migrations to rollback")

	// Mark as required
	migrateCreateCmd.MarkFlagRequired("name")
	migrateRollbackCmd.MarkFlagRequired("count")

	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateSchemaCmd)
	migrateCmd.AddCommand(migrateSQLApply)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateRehashCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCommand)

	DBCMd.AddCommand(migrateCmd)
}
