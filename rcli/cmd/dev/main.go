// Package devcmd
package devcmd

import (
	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development commands",
}

var devUpCommand = &cobra.Command{
	Use:     "start",
	Aliases: []string{"run", "up"},
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("docker", "compose", "up", "--build")
	},
}

func init() {
	DevCmd.AddCommand(devUpCommand)
}
