// Package testcmd
package testcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var config *utils.Environment

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands",
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup test environment",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsServiceDir() {
			panic("Not in a service directory")
		}
		fmt.Println("🔧 Setting up test environment...")

		pool := utils.MustConnectDB("localhost", 5433, "test_user", "test_password", "test_db", "disable")

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

var testUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate database",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("docker", "compose", "-f", "docker-compose.test.yml", "up", "--build")
	},
}

func init() {
	if utils.IsServiceDir() {
		config = utils.MustLoadEnv()
	}

	TestCmd.AddCommand(testSetupCmd)
	TestCmd.AddCommand(testRunCmd)
	TestCmd.AddCommand(testUpCmd)
}
