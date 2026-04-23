package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func CopyFile(source, destination string) error {
	sourceFilename := filepath.Base(source)
	destFilename := filepath.Base(destination)

	sourceBytes, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", sourceFilename, err)
	}

	destDir := filepath.Dir(destination)
	// Skip if file already exists
	if _, err := os.Stat(destination); err == nil {
		fmt.Printf("⏭️  Skipping %s; it already exists in %s\n", destFilename, destDir)
		return nil
	}

	if err := os.WriteFile(destination, sourceBytes, 0o644); err != nil {
		return fmt.Errorf("write %s to %s: %w", destFilename, destDir, err)
	}

	return nil
}

func IsServiceDir() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Check if current directory name is "services" or parent is "services"
	parts := strings.Split(cwd, string(os.PathSeparator))
	for i, part := range parts {
		if part == "services" && i < len(parts)-1 {
			// We're inside a services directory (not the services dir itself)
			return true
		}
	}
	return false
}

func GetRootDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		// Check for project markers
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd, nil
}

func IsRootDir() bool {
	if _, err := os.Stat("go.mod"); err != nil {
		return false
	}
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return false
	}

	if strings.Contains(string(content), "module github.com/rijum8906/relay/") {
		return false
	}

	return strings.Contains(string(content), "module github.com/rijum8906/relay")
}
