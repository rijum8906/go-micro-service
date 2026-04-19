package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run task service setup actions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No setup actions are required for the task-service skeleton.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
