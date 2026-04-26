package cmd

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/command"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Manage the local test environment",
	Long:  "Manage the local PostgreSQL-backed test environment for the task service.",
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare test environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Setting up test environment...")

		fmt.Println("Cleaning test database...")
		if err := runCommand(
			"atlas", "schema", "clean",
			"--url", testDBURL(),
			"--auto-approve",
		); err != nil {
			return fmt.Errorf("clean test database: %w", err)
		}

		fmt.Println("Applying test migrations...")
		if err := runCommand(
			"atlas", "migrate", "apply",
			"--url", testDBURL(),
			"--dir", migrationDir(),
		); err != nil {
			return fmt.Errorf("apply test migrations: %w", err)
		}

		fmt.Println("Test environment setup complete.")
		return nil
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

func testDBURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		testutils.DBUser,
		testutils.DBPassword,
		testutils.DBHost,
		testutils.DBPort,
		testutils.DBName,
		testutils.DBSSLMode,
	)
}
