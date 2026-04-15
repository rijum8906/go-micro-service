package cmd

import (
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
	Run:   command.NotImplemented,
}

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests",
	Run:   command.NotImplemented,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start test dependencies",
	Run:   command.NotImplemented,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop test dependencies",
	Run:   command.NotImplemented,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run test migrations",
	Run:   command.NotImplemented,
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all test actions",
	Run:   command.NotImplemented,
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.AddCommand(testSetupCmd)
	testCmd.AddCommand(testRunCmd)
	testCmd.AddCommand(startCmd)
	testCmd.AddCommand(stopCmd)
	testCmd.AddCommand(migrateCmd)
	testCmd.AddCommand(allCmd)
}
