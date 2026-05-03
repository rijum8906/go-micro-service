package projectcmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var useAtlas bool

var initDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Initialize test database and development database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n🚀 Setting up databases")
		servics, err := utils.GetServices()
		if err != nil {
			return err
		}

		for _, svc := range servics {
			fmt.Println(svc)
			cmd1 := exec.Command("rcli", "db", "init")
			cmd1.Stdout = os.Stdout
			cmd1.Stderr = os.Stderr
			cmd1.Dir = svc
			if err := cmd1.Run(); err != nil {
				continue
			}

			if useAtlas {
				cmd2 := exec.Command("rcli", "db", "migrate", "apply")
				cmd2.Dir = svc
				cmd2.Stdout = os.Stdout
				cmd2.Stderr = os.Stderr
				if err := cmd2.Run(); err != nil {
					continue
				}

			} else {
				cmd2 := exec.Command("rcli", "db", "exec", "apply")
				cmd2.Dir = svc
				cmd2.Stdout = os.Stdout
				cmd2.Stderr = os.Stderr
				if err := cmd2.Run(); err != nil {
					continue
				}
			}
		}

		return nil
	},
}

func init() {
	initDBCmd.Flags().BoolVarP(&useAtlas, "use-atlas", "a", false, "Use atlas database")
	initProjectCmd.AddCommand(initDBCmd)
}
