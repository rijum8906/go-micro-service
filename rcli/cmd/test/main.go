// Package testcmd
package testcmd

import (
	"fmt"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands",
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup test environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsServiceDir() {
			return fmt.Errorf("must be run from a service directory")
		}
		fmt.Println("\n🧪 Setting up test environment")

		return nil
	},
}

var testUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.RunCommand("docker", "compose", "-f", "docker-compose.test.yml", "up", "--build")
	},
}

func init() {
	TestCmd.AddCommand(testSetupCmd)
	TestCmd.AddCommand(testRunCmd)
	TestCmd.AddCommand(testUpCmd)
}
