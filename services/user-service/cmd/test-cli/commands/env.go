package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	coredb "github.com/rijum8906/relay/packages/core/db"
	"github.com/rijum8906/relay/packages/core/testutils"
	migrations "github.com/rijum8906/relay/services/user/db"
)

func (c *CLI) StartAll() error {
	fmt.Fprintln(c.Stdout, "Starting local test containers...")

	for _, container := range c.ContainerManager.Containers {
		if container.Exists() {
			if container.IsRunning() {
				fmt.Fprintf(c.Stdout, "Container already running: %s\n", container.Name)
				continue
			}

			fmt.Fprintf(c.Stdout, "Removing stale container: %s\n", container.Name)
			if err := container.Remove(); err != nil {
				return fmt.Errorf("remove stale container %s: %w", container.Name, err)
			}
		}

		fmt.Fprintf(c.Stdout, "Starting container: %s\n", container.Name)
		if err := container.Run(); err != nil {
			return fmt.Errorf("start container %s: %w", container.Name, err)
		}
	}

	postgres := c.ContainerManager.GetContainer("user-service-test-postgres")
	if postgres != nil {
		fmt.Fprintln(c.Stdout, "Waiting for PostgreSQL to become ready...")
		if err := postgres.WaitForReady(30*time.Second, c.isPostgresReady); err != nil {
			return err
		}
	}

	redis := c.ContainerManager.GetContainer("user-service-test-redis")
	if redis != nil {
		fmt.Fprintln(c.Stdout, "Waiting for Redis to become ready...")
		if err := redis.WaitForReady(15*time.Second, redis.IsRunning); err != nil {
			return err
		}
	}

	fmt.Fprintln(c.Stdout, "Local test environment is ready.")
	return nil
}

func (c *CLI) StopAll() error {
	fmt.Fprintln(c.Stdout, "Stopping local test containers...")
	if err := c.ContainerManager.StopAll(); err != nil {
		return fmt.Errorf("stop local test containers: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Containers stopped.")
	return nil
}

func (c *CLI) MigrateTestDB() error {
	postgres := c.ContainerManager.GetContainer("user-service-test-postgres")
	if postgres == nil || !postgres.Exists() {
		return fmt.Errorf("test postgres container is not running; start it with `env-start` first")
	}

	ctx := context.Background()
	pool, appErr := coredb.Connect(ctx, testutils.DBConf)
	if appErr != nil {
		return fmt.Errorf("connect test database: %w", appErr)
	}
	defer pool.Close()

	allMigrations, err := migrations.All()
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	fmt.Fprintln(c.Stdout, "Applying embedded SQL migrations to the test database...")
	for _, migration := range allMigrations {
		statements := splitSQLStatements(migration.Content)
		for _, statement := range statements {
			if _, err := pool.Exec(ctx, statement); err != nil {
				if shouldIgnoreMigrationError(err) {
					continue
				}
				return fmt.Errorf("apply test migration %s: %w", migration.Name, err)
			}
		}
	}

	fmt.Fprintln(c.Stdout, "Test database migrated.")
	return nil
}

func (c *CLI) isPostgresReady() bool {
	postgres := c.ContainerManager.GetContainer("user-service-test-postgres")
	if postgres == nil {
		return false
	}

	return execCheck(
		"docker", "exec", postgres.Name, "pg_isready", "-U", testutils.DBUser,
	)
}

func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement == "" {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func shouldIgnoreMigrationError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "42P07", "42710":
		return true
	default:
		return false
	}
}
