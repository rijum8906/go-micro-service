package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/rijum8906/relay/cli/utils"
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage local Relay services",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsServiceDir() {
			return fmt.Errorf("current directory is not a service directory")
		}

		return nil
	},
}

// #################################################
//              DEV COMMANDS
// #################################################

var serviceDevCmd = &cobra.Command{
	Use:     "dev",
	Aliases: []string{"development"},
}

var devSetupCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"init"},
	Run: func(cmd *cobra.Command, args []string) {
		currentOS := runtime.GOOS
		fmt.Printf("⚙️ Running setup for %s\n", currentOS)

		// Step 1: Install sqlc
		installGoPackage("sqlc", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest")

		// Step 2: Install Atlas
		if currentOS == "linux" || currentOS == "darwin" {
			installCurlBinary("atlas", "https://atlasgo.sh")
		} else {
			fmt.Printf("⚠️ No installation command for %s, install manually\n", currentOS)
		}

		// Done
		fmt.Println("🎉 Setup complete")
	},
}

var devStartCommand = &cobra.Command{
	Use:     "start",
	Aliases: []string{"run", "up"},
	Run: func(cmd *cobra.Command, args []string) {
		runCommand("docker", "compose", "up", "--build")
	},
}

// #################################################
//              TEST COMMANDS
// #################################################

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands",
}

var testSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare test environment",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting up test environment...")
		testSetupDBCmd.Run(cmd, args)
	},
}

var testSetupDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Prepare test database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Migrating test database...")

		pool := utils.MustConnectDB()

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
		runCommand("go", "test", "./...")
	},
}

func init() {
	serviceDevCmd.AddCommand(devSetupCmd)
	serviceDevCmd.AddCommand(devStartCommand)

	testSetupCmd.AddCommand(testSetupDBCmd)

	testCmd.AddCommand(testSetupCmd)
	testCmd.AddCommand(testRunCmd)

	serviceCmd.AddCommand(serviceDevCmd)
	serviceCmd.AddCommand(testCmd)

	mainCmd.AddCommand(serviceCmd)
}
