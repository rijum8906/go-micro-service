// Package cmd
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// mainCmd represents the base command when called without any subcommands
var mainCmd = &cobra.Command{
	Use:   "my-cli",
	Short: "Manage local Relay development workflows",
	Long: `my-cli provides a small command surface for common Relay project tasks.

Use it to set up the workspace, prepare local environment files, and run
development services without repeating the underlying Go and Docker commands.`,
}

func Execute() {
	err := mainCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
