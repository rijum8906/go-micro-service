/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the local development stack",
	Long: `dev builds and starts the local Relay services with Docker Compose.

Use this command when you want to run the project in development mode and
rebuild containers from the current source before startup.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand("docker", "compose", "up", "--build")
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
