package cmd

import (
	"github.com/rijum8906/relay/services/task-service/app"
	"github.com/spf13/cobra"
)

// serviceRunCmd represents the run command
var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the task service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.RunService(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serviceRunCmd)
}
