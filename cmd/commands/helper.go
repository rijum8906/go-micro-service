package commands

import (
	"fmt"
	"os"
	"os/exec"
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
