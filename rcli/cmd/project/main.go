// Package projectcmd
package projectcmd

import (
	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project commands",
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the project",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("go", "work", "use", "./services/*")
	},
}

func init() {
	ProjectCmd.AddCommand(syncCmd)
}
