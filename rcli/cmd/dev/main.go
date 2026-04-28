// Package devcmd
package devcmd

import (
	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var runInfra bool

var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development commands",
}

var devUpCommand = &cobra.Command{
	Use:     "start",
	Aliases: []string{"run", "up"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if runInfra {
			return utils.RunCommand("docker", "compose", "up", "nats", "postgres", "redis")
		}
		return utils.RunCommand("docker", "compose", "up")
	},
}

func init() {
	devUpCommand.Flags().BoolVarP(&runInfra, "infra", "i", false, "Run infra services")
	DevCmd.AddCommand(devUpCommand)
}
