package cmd

import (
	"github.com/spf13/cobra"
)

// serviceRunCmd represents the run command
var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run notification service actions",
	Run:   notImplemented,
}

func init() {
	rootCmd.AddCommand(serviceRunCmd)
}
