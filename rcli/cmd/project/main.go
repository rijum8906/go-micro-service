// Package projectcmd
package projectcmd

import (
	"fmt"
	"path/filepath"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project commands",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsRootDir() {
			return fmt.Errorf("must run from root directory")
		}

		return nil
	},
}

var initProjectCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"setup"},
	Short:   "Initialize the Relay project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n🚀 Setting up project")

		// Step 1: Initialize Go workspace
		if err := utils.RunCommand("go", "work", "init"); err != nil {
			return err
		}

		// Step 2: Add services and packages (cross-platform)
		if err := addWorkspaceDirectories(); err != nil {
			return err
		}

		// Step 3: Download dependencies
		if err := utils.RunCommand("go", "work", "sync"); err != nil {
			return err
		}
		if err := utils.RunCommand("go", "mod", "download"); err != nil {
			return err
		}

		// Step 4: Copy .env files
		fmt.Println("\n📄 Copying environment files to services")
		if err := copyEnvFiles(); err != nil {
			return err
		}

		fmt.Println("\n✅ Setup complete")
		return nil
	},
}

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n🔄 Syncing project")

		if err := addWorkspaceDirectories(); err != nil {
			return err
		}
		if err := utils.RunCommand("go", "work", "sync"); err != nil {
			return err
		}
		if err := utils.RunCommand("go", "mod", "download"); err != nil {
			return err
		}
		if err := copyEnvFiles(); err != nil {
			return err
		}

		fmt.Println("\n✅ Sync complete")
		return nil
	},
}

func addWorkspaceDirectories() error {
	fmt.Println("\n📦 Adding directories to workspace")

	// Add root
	if err := utils.RunCommand("go", "work", "use", "."); err != nil {
		return err
	}

	// Add all packages
	packages, err := filepath.Glob("packages/*")
	if err != nil {
		return fmt.Errorf("find packages: %w", err)
	} else {
		for _, pkg := range packages {
			if utils.IsDirectory(pkg) {
				if err := utils.RunCommand("go", "work", "use", pkg); err != nil {
					return fmt.Errorf("add package %s to workspace: %w", pkg, err)
				}
			}
		}
	}

	// Add all services
	services, err := filepath.Glob("services/*")
	if err != nil {
		return fmt.Errorf("find services: %w", err)
	} else {
		for _, svc := range services {
			if utils.IsDirectory(svc) {
				if err := utils.RunCommand("go", "work", "use", svc); err != nil {
					return fmt.Errorf("add service %s to workspace: %w", svc, err)
				}
			}
		}
	}

	return nil
}

func copyEnvFiles() error {
	services, err := utils.GetServices()
	if err != nil {
		return err
	}

	for _, svc := range services {
		if err := utils.CopyFile(filepath.Join(svc, ".env.example"), filepath.Join(svc, ".env")); err != nil {
			return fmt.Errorf("copy environment file for %s: %w", svc, err)
		}
	}

	return nil
}

func init() {
	ProjectCmd.AddCommand(initProjectCmd)
	ProjectCmd.AddCommand(projectSyncCmd)
}
