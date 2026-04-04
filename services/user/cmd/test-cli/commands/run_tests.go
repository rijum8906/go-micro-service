package commands

import (
	"os"
	"os/exec"
)

func (c *CLI) RunTests() {
	println("Running tests...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		println("Tests failed:", err.Error())
		os.Exit(1)
	}
	println("Tests passed!")
}
