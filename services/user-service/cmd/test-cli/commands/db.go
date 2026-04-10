package commands

import "fmt"

func (c *CLI) Setup() error {
	fmt.Fprintln(c.Stdout, "Installing sqlc...")
	if err := c.runCommand("go", "install", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest"); err != nil {
		return fmt.Errorf("install sqlc: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Installing atlas...")
	if err := c.runShell("curl -sSf https://atlasgo.sh | sh"); err != nil {
		return fmt.Errorf("install atlas: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Setup complete.")
	return nil
}

func (c *CLI) DBNew(name string) error {
	fmt.Fprintf(c.Stdout, "Generating migration: %s\n", name)
	if err := c.runCommand(
		"atlas", "migrate", "diff", name,
		"--dir", c.migrationDir(),
		"--to", c.schemaURL(),
		"--dev-url", c.devDBURL(),
	); err != nil {
		return fmt.Errorf("generate migration: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Migration generated.")
	return nil
}

func (c *CLI) DBApply() error {
	fmt.Fprintln(c.Stdout, "Applying migrations...")
	if err := c.runCommand(
		"atlas", "migrate", "apply",
		"--url", c.dbURL(),
		"--dir", c.migrationDir(),
	); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Migrations applied.")
	return nil
}

func (c *CLI) DBStatus() error {
	fmt.Fprintln(c.Stdout, "Checking migration status...")
	return c.runCommand(
		"atlas", "migrate", "status",
		"--url", c.dbURL(),
		"--dir", c.migrationDir(),
	)
}

func (c *CLI) DBSchema() error {
	fmt.Fprintln(c.Stdout, "Applying schema...")
	if err := c.runCommand(
		"atlas", "schema", "apply",
		"--url", c.dbURL(),
		"--to", c.schemaURL(),
		"--dev-url", c.devDBURL(),
		"--auto-approve",
	); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Schema applied.")
	return nil
}

func (c *CLI) DBRehash() error {
	fmt.Fprintln(c.Stdout, "Rehashing migration directory...")
	if err := c.runCommand("atlas", "migrate", "hash", "--dir", c.migrationDir()); err != nil {
		return fmt.Errorf("rehash migrations: %w", err)
	}

	fmt.Fprintln(c.Stdout, "atlas.sum updated.")
	return nil
}

func (c *CLI) DBRollback(count int) error {
	fmt.Fprintf(c.Stdout, "Rolling back %d migration(s)...\n", count)
	if err := c.runCommand(
		"atlas", "migrate", "down", fmt.Sprint(count),
		"--url", c.dbURL(),
		"--dir", c.migrationDir(),
		"--dev-url", c.devDBURL(),
	); err != nil {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Rollback complete.")
	return nil
}
