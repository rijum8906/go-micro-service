package devcmd

import (
	"fmt"
	"runtime"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var devInstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentOS := runtime.GOOS
		fmt.Printf("\n🔧 Running setup for %s\n", currentOS)

		// Step 1: Install sqlc
		if err := utils.InstallGoPackage("sqlc", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest"); err != nil {
			return err
		}

		// Step 2: Install Atlas
		if currentOS == "linux" || currentOS == "darwin" {
			if err := utils.InstallCurlBinary("atlas", "https://atlasgo.sh"); err != nil {
				return err
			}
		} else {
			fmt.Printf("\n⚠️  No installation command for %s; install Atlas manually\n", currentOS)
		}

		// Done
		fmt.Println("\n✅ Setup complete")
		return nil
	},
}

func init() {
	DevCmd.AddCommand(devInstallCommand)
}
