package cmd

import (
	"fmt"
	"runtime"

	"github.com/rijum8906/relay/packages/core/command"
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
		command.InstallGoPackage("sqlc", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest")

		// Step 2: Install Atlas
		if currentOS == "linux" || currentOS == "darwin" {
			command.InstallCurlBinary("atlas", "https://atlasgo.sh")
		} else {
			fmt.Printf("⚠️ No installation command for %s, install manually\n", currentOS)
		}

		// Done
		fmt.Println("🎉 Setup complete")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
