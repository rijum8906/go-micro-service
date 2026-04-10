package cmd

import (
	"fmt"

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

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare the local test environment",
	Long:  "Prepare the local PostgreSQL and Redis test containers for local development and testing.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if testContainerManager.ExistsAll() {
			fmt.Println("Local test containers already exist. Refreshing them.")
		}

		return startContainers()
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local test containers",
	Long:  "Start the local PostgreSQL and Redis containers used by the test environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startContainers()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local test containers",
	Long:  "Stop the local PostgreSQL and Redis containers used by the test environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopContainers()
	},
}

func init() {
	testCmd.AddCommand(runCmd)
	testCmd.AddCommand(setupCmd)
	testCmd.AddCommand(startCmd)
	testCmd.AddCommand(stopCmd)

	rootCmd.AddCommand(testCmd)
}
