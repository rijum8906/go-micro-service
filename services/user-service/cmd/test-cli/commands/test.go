package commands

import (
	"fmt"
	"os/exec"
)

func (c *CLI) RunTests() error {
	fmt.Fprintln(c.Stdout, "Running Go tests...")
	if err := c.runCommand("go", "test", "-v", "./..."); err != nil {
		return fmt.Errorf("run tests: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Tests passed.")
	return nil
}

func (c *CLI) RunIntegrationFlow() error {
	if err := c.StartAll(); err != nil {
		return err
	}

	if err := c.MigrateTestDB(); err != nil {
		return err
	}

	return c.RunTests()
}

func execCheck(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	return cmd.Run() == nil
}
