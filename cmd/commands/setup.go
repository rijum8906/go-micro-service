// Package commands
package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

var Services = map[string]string{
	"user-service":    "services/user-service/",
	"graphql-gateway": "services/graphql-gateway/",
}

func Setup() {
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
			if isDirectory(pkg) {
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
			if isDirectory(svc) {
				runCommand("go", "work", "use", svc)
			}
		}
	}
}

func copyEnvFiles() {
	fmt.Println("\n📦 Copying .env files...")

	source, err := os.ReadFile(".env.example")
	if err != nil {
		panic(fmt.Errorf("failed to read .env.example: %w", err))
	}

	for name, path := range Services {
		destPath := filepath.Join(path, ".env")

		// Skip if .env already exists
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("  ⚠️  .env already exists in %s, skipping\n", name)
			continue
		}

		if err := os.WriteFile(destPath, source, 0o644); err != nil {
			panic(fmt.Errorf("failed to write .env to %s: %w", name, err))
		}
		fmt.Printf("  ✅ Copied .env to %s\n", name)
	}
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
