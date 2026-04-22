// Package initcmd
package initcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rijum8906/relay/cli/utils"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Relay project or service",
}

var initProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Initialize the Relay project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Setting up project...")

		// Step 1: Initialize Go workspace
		utils.RunCommand("go", "work", "init")

		// Step 2: Add services and packages (cross-platform)
		addWorkspaceDirectories()

		// Step 3: Download dependencies
		utils.RunCommand("go", "work", "sync")
		utils.RunCommand("go", "mod", "download")

		// Step 4: Copy .env files
		copyEnvFiles()

		fmt.Println("\n🎉 Setup complete!")
	},
}

// addWorkspaceDirectories finds all packages and services and adds them to workspace
func addWorkspaceDirectories() {
	fmt.Println("\n📦 Adding directories to workspace...")

	// Add root
	utils.RunCommand("go", "work", "use", ".")

	// Add all packages
	packages, err := filepath.Glob("packages/*")
	if err != nil {
		fmt.Printf("  ⚠️  Failed to find packages: %v\n", err)
	} else {
		for _, pkg := range packages {
			if utils.IsDirectory(pkg) {
				utils.RunCommand("go", "work", "use", pkg)
			}
		}
	}

	// Add all services
	services, err := filepath.Glob("services/*")
	if err != nil {
		fmt.Printf("  ⚠️  Failed to find services: %v\n", err)
	} else {
		for _, svc := range services {
			if utils.IsDirectory(svc) {
				utils.RunCommand("go", "work", "use", svc)
			}
		}
	}
}

func copyEnvFiles() {
	services, err := getServices()
	if err != nil {
		panic(err)
	}

	for _, svc := range services {
		utils.CopyFile(filepath.Join("services", svc, ".env.example"), filepath.Join("services", svc, ".env"))
	}
}

func getServices() ([]string, error) {
	entries, err := os.ReadDir("services")
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	services := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			services = append(services, entry.Name())
		}
	}

	return services, nil
}

var initServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Initialize the Relay service",
	Run:   utils.NotImplemented,
}

func init() {
	InitCmd.AddCommand(initProjectCmd)
	InitCmd.AddCommand(initServiceCmd)
}
