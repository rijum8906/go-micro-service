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

	// Step 2: Add services and packages (using shell for wildcards)
	runCommand("sh", "-c", "go work use ./packages/*")
	runCommand("sh", "-c", "go work use .")
	runCommand("sh", "-c", "go work use ./services/*")

	// Step 3: Download dependencies
	runCommand("go", "work", "sync")
	runCommand("go", "mod", "download")

	// Step 4: Copy .env files
	fmt.Println("\n📦 Copying .env files...")
	source, err := os.ReadFile(".env.example")
	if err != nil {
		panic(fmt.Errorf("failed to read .env.example: %w", err))
	}

	for name, path := range Services {
		destPath := filepath.Join(path, ".env")
		if err := os.WriteFile(destPath, source, 0o644); err != nil {
			panic(fmt.Errorf("failed to write .env to %s: %w", name, err))
		}
		fmt.Printf("  ✅ Copied .env to %s\n", name)
	}

	fmt.Println("\n🎉 Setup complete!")
}
