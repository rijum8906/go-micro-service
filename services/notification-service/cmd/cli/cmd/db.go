package cmd

import (
	"fmt"

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

var atlasDBInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize atlas database",
	Run: func(cmd *cobra.Command, args []string) {
		defaultCfg := *config
		defaultCfg.DBName = "postgres"
		err := command.CreateNewDatabase(&defaultCfg, config.DBName)
		if err != nil {
			fmt.Printf("failed to create database %s: %v", config.DBName, err)
		}

		err = command.CreateNewDatabase(&defaultCfg, "dev_"+config.DBName)
		if err != nil {
			fmt.Printf("failed to create database %s: %v", "dev_"+config.DBName, err)
		}
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
