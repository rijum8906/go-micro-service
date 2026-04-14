package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run notification service setup actions",
	Run: func(cmd *cobra.Command, args []string) {
		currentOS := runtime.GOOS
		fmt.Printf("⚙️ Running setup for %s\n", currentOS)

		// Step 1: Install sqlc
		if !isCommandAvailable("sqlc") {
			fmt.Println("📦 Installing sqlc...")
			runCommand("go", "install", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest")
		} else {
			fmt.Println("✅ sqlc already installed")
		}

		// Step 2: Install Atlas
		if !isCommandAvailable("atlas") {
			if currentOS == "linux" || currentOS == "darwin" {
				fmt.Println("📦 Installing atlas...")
				runCommand("sh", "-c", "curl -sSf https://atlasgo.sh | sh")
			} else {
				fmt.Printf("⚠️ No installation command for %s, install manually\n", currentOS)
			}
		} else {
			fmt.Println("✅ atlas already installed")
		}

		// Done
		fmt.Println("🎉 Setup complete")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
