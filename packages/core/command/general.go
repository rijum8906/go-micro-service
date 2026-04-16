// Package command
package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func RunCommand(name string, args ...string) {
	fmt.Printf("  → Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("command failed: %s %v: %w", name, args, err))
	}
}

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NotImplemented(_ *cobra.Command, _ []string) {
	fmt.Println("🚧 This command is not implemented yet.")
}
