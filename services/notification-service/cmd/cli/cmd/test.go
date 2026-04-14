package cmd

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands",
	Run:   notImplemented,
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare test environment",
	Run:   notImplemented,
}

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests",
	Run:   notImplemented,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start test dependencies",
	Run:   notImplemented,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop test dependencies",
	Run:   notImplemented,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run test migrations",
	Run:   notImplemented,
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all test actions",
	Run:   notImplemented,
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
