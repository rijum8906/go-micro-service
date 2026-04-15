package cmd

import (
	"github.com/rijum8906/relay/packages/core/command"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/spf13/cobra"
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

var atlasApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply atlas migrations",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas",
			"migrate", "apply",
			"--url", getDBURL(config),
			"--dir", getMigrationDir())
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show atlas status",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas",
			"migrate", "status",
			"--url", getDBURL(config),
			"--dir", getMigrationDir())
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new atlas migration",
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		command.RunCommand("atlas", "migrate", "diff", migrationName,
			"--dir", getMigrationDir(),
			"--to", getSchemaDir(),
			"--dev-url", getDevDBURL(config))
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Apply schema actions",
	Run: func(cmd *cobra.Command, args []string) {
		err := createNewDatabase(config, "dev_"+config.DBName)
		if err != nil {
			panic(err)
		}
		command.RunCommand("atlas",
			"schema", "apply",
			"--url", getDBURL(config),
			"--dev-url", getDevDBURL(config),
			"--file", getSchemaDir(),
			"--auto-approve")
	},
}

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rehash atlas migration checksums",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("atlas", "migrate", "hash",
			"--dir", getMigrationDir())
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback atlas migrations",
	Run: func(cmd *cobra.Command, args []string) {
		count := args[0]
		command.RunCommand("atlas", "migrate", "down", count,
			"--url", getDBURL(config),
			"--dir", getMigrationDir(),
			"--dev-url", getDevDBURL(config))
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
	Run:   command.NotImplemented,
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

	atlasCmd.AddCommand(atlasApplyCmd)
	atlasCmd.AddCommand(statusCmd)
	atlasCmd.AddCommand(newCmd)
	atlasCmd.AddCommand(schemaCmd)
	atlasCmd.AddCommand(rehashCmd)
	atlasCmd.AddCommand(rollbackCmd)
	atlasCmd.AddCommand(resetCmd)

	sqlCmd.AddCommand(sqlApplyCmd)
}
