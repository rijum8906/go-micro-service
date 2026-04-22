// Package cmd
package cmd

import (
	"os"

	dbcmd "github.com/rijum8906/relay/rcli/cmd/db"
	devcmd "github.com/rijum8906/relay/rcli/cmd/dev"
	initcmd "github.com/rijum8906/relay/rcli/cmd/init"
	testcmd "github.com/rijum8906/relay/rcli/cmd/test"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "my-cli",
	Short: "Manage local Relay development workflows",
	Long: `my-cli provides a small command surface for common Relay project tasks.

Use it to set up the workspace, prepare local environment files, and run
development services without repeating the underlying Go and Docker commands.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initcmd.InitCmd)
	rootCmd.AddCommand(devcmd.DevCmd)
	rootCmd.AddCommand(dbcmd.DBCMd)
	rootCmd.AddCommand(testcmd.TestCmd)
}
