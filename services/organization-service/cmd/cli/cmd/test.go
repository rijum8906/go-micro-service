package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rijum8906/relay/packages/core/command"
	"github.com/rijum8906/relay/packages/core/testutils"
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
		fmt.Println("Setting up test environment...")

		fmt.Println("Migrating test database...")

		pool := testutils.MustConnectDB()

		content, err := os.ReadFile("db/schema.sql")
		if err != nil {
			panic(err)
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			fmt.Printf("failed to apply migration %s: %v", "db/schema.sql", err)
		}

		fmt.Println("Test environment setup complete.")
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
