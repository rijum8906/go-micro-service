package cmd

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/command"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands",
	Run:   command.NotImplemented,
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare test environment",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No test environment setup is required for the task-service skeleton.")
	},
}

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests",
	Run: func(cmd *cobra.Command, args []string) {
		command.RunCommand("go", "test", "-v", "./...")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.AddCommand(testSetupCmd)
	testCmd.AddCommand(testRunCmd)
}
