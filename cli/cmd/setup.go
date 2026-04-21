package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rijum8906/relay/cli/utils"
	"github.com/spf13/cobra"
)

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

// ############################################
//              ROOT COMMANDS
// ############################################

var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "Root commands",
	Run:   notImplemented,
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare the Relay workspace for local development",
	Long: `setup initializes the Go workspace, registers local modules, syncs
dependencies, and creates service environment files.

Run this once after cloning the repository or whenever the workspace layout
changes and you need to regenerate local setup artifacts.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Setting up project...")

		// Step 1: Initialize Go workspace
		runCommand("go", "work", "init")

		// Step 2: Add services and packages (cross-platform)
		addWorkspaceDirectories()

		// Step 3: Download dependencies
		runCommand("go", "work", "sync")
		runCommand("go", "mod", "download")

		// Step 4: Copy .env files
		copyEnvFiles()

		fmt.Println("\n🎉 Setup complete!")
	},
}

// addWorkspaceDirectories finds all packages and services and adds them to workspace
func addWorkspaceDirectories() {
	fmt.Println("\n📦 Adding directories to workspace...")

	// Add root
	runCommand("go", "work", "use", ".")

	// Add all packages
	packages, err := filepath.Glob("packages/*")
	if err != nil {
		fmt.Printf("  ⚠️  Failed to find packages: %v\n", err)
	} else {
		for _, pkg := range packages {
			if utils.IsDirectory(pkg) {
				runCommand("go", "work", "use", pkg)
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
				runCommand("go", "work", "use", svc)
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

func init() {
	rootCmd.AddCommand(setupCmd)

	mainCmd.AddCommand(rootCmd)
}
