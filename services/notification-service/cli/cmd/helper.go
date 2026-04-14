package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runCommand(name string, args ...string) {
	fmt.Printf("  → Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("command failed: %s %v: %w", name, args, err))
	}
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func copyFile(source, destination string) {
	sourceFilename := filepath.Base(source)
	destFilename := filepath.Base(destination)

	fmt.Printf("\n📦 Copying %s file to %s...\n", sourceFilename, destFilename)

	sourceBytes, err := os.ReadFile(source)
	if err != nil {
		panic(fmt.Errorf("failed to read %s: %w", sourceFilename, err))
	}

	destDir := filepath.Dir(destination)
	// Skip if file already exists
	if _, err := os.Stat(destination); err == nil {
		fmt.Printf("  ⚠️  %s already exists in %s, skipping\n", destFilename, destDir)
		return
	}

	if err := os.WriteFile(destination, sourceBytes, 0o644); err != nil {
		panic(fmt.Errorf("failed to write %s to %s: %w", destFilename, destDir, err))
	}
	fmt.Printf("  ✅ Copied %s to %s\n", destFilename, destDir)
}
