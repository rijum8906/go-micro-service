package cmd

import (
	"github.com/rijum8906/relay/packages/core/command"
	"github.com/spf13/cobra"
)

// serviceRunCmd represents the run command
var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run task service actions",
	Run:   command.NotImplemented,
}

func init() {
	rootCmd.AddCommand(serviceRunCmd)
}
