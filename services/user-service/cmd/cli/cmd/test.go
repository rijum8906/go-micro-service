package cmd

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/testutils"
	migrations "github.com/rijum8906/relay/services/user/db"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Manage the local test environment",
	Long:  "Manage the Docker containers used for the user service local test environment.",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Ensure the local test containers are running",
	Long:  "Ensure the local PostgreSQL and Redis test containers are available and running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if testContainerManager.ExistsAll() {
			fmt.Println("Local test containers already exist. Starting them again to ensure a clean state.")
		}

		return startContainers()
	},
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare test environment",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting up test environment...")

		fmt.Println("Migrating test database...")

		pool := testutils.MustConnectDB()

		migrations, err := migrations.All()
		if err != nil {
			panic(err)
		}

		for _, m := range migrations {
			_, err = pool.Exec(context.Background(), m.Content)
			if err != nil {
				fmt.Printf("failed to apply migration %s: %v", m.Name, err)
			}
		}

		fmt.Println("Test environment setup complete.")
	},
}

func init() {
	testCmd.AddCommand(runCmd)
	testCmd.AddCommand(testSetupCmd)

	rootCmd.AddCommand(testCmd)
}
