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
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.RunCommand("docker", "compose", "up", "--build")
	},
}

func init() {
	DevCmd.AddCommand(devUpCommand)
}
