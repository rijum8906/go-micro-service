package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func RunCommand(name string, args ...string) error {
	command := formatCommandForLog(name, args)
	fmt.Printf("\n▶️  Running: %s\n", command)

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed (%s): %w", command, err)
	}

	return nil
}

func RunCommandSilent(name string, args ...string) error {
	command := formatCommandForLog(name, args)
	fmt.Printf("\n▶️  Running: %s\n", command)

	cmd := exec.Command(name, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed (%s): %w", command, err)
	}

	return nil
}

func formatCommandForLog(name string, args []string) string {
	parts := []string{name}
	redactNext := false

	for _, arg := range args {
		if redactNext {
			parts = append(parts, "<redacted>")
			redactNext = false
			continue
		}

		if arg == "--url" || arg == "--dev-url" {
			parts = append(parts, arg)
			redactNext = true
			continue
		}

		if strings.HasPrefix(arg, "--url=") || strings.HasPrefix(arg, "--dev-url=") {
			parts = append(parts, strings.SplitN(arg, "=", 2)[0]+"=<redacted>")
			continue
		}

		if strings.HasPrefix(arg, "postgres://") || strings.HasPrefix(arg, "postgresql://") {
			parts = append(parts, "<redacted>")
			continue
		}

		parts = append(parts, arg)
	}

	return strings.Join(parts, " ")
}

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NotImplemented(_ *cobra.Command, _ []string) {
	fmt.Println("\n🚧 This command is not implemented yet.")
}

func InstallGoPackage(name, url string) error {
	if IsCommandAvailable(name) {
		fmt.Printf("\n⏭️  %s is already installed\n", name)
		return nil
	}

	fmt.Printf("\n📦 Installing %s with Go\n", name)
	if err := RunCommand("go", "install", url); err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}

	return nil
}

func InstallCurlBinary(name, url string) error {
	if IsCommandAvailable(name) {
		fmt.Printf("\n⏭️  %s is already installed\n", name)
		return nil
	}

	fmt.Printf("\n📦 Installing %s\n", name)
	if err := RunCommand("sh", "-c", "curl -sSf "+url+" | sh"); err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}

	return nil
}
