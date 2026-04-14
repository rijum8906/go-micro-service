package cmd

import "github.com/spf13/cobra"

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
	Run:   notImplemented,
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
