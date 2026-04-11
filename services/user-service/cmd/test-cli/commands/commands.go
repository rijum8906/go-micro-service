package commands

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/rijum8906/relay/packages/core/testutils"
)

const (
	defaultDBName     = "postgres"
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBHost     = "localhost"
	defaultDBPort     = 5432
	defaultDBSSLMode  = "disable"
)

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		c.printUsage(c.Stderr)
		return errors.New("command is required")
	}

	switch args[0] {
	case "help", "-h", "--help":
		c.printUsage(c.Stdout)
		return nil
	case "setup":
		return c.Setup()
	case "db-new":
		return c.runDBNew(args[1:])
	case "db-apply":
		return c.DBApply()
	case "db-status":
		return c.DBStatus()
	case "db-schema":
		return c.DBSchema()
	case "db-rehash":
		return c.DBRehash()
	case "db-rollback":
		return c.runDBRollback(args[1:])
	case "test":
		return c.RunTests()
	case "env-start":
		return c.StartAll()
	case "env-stop":
		return c.StopAll()
	case "env-migrate":
		return c.MigrateTestDB()
	case "env-test":
		return c.RunIntegrationFlow()
	default:
		c.printUsage(c.Stderr)
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (c *CLI) printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  go run ./cmd/test-cli <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  setup                   Install development tools (sqlc, atlas)")
	fmt.Fprintln(w, "  db-new --name <name>    Generate a new Atlas migration")
	fmt.Fprintln(w, "  db-apply                Apply pending migrations")
	fmt.Fprintln(w, "  db-status               Show migration status")
	fmt.Fprintln(w, "  db-schema               Apply schema directly")
	fmt.Fprintln(w, "  db-rehash               Recalculate atlas.sum")
	fmt.Fprintln(w, "  db-rollback --count N   Roll back migrations")
	fmt.Fprintln(w, "  test                    Run go test")
	fmt.Fprintln(w, "  env-start               Start local test containers")
	fmt.Fprintln(w, "  env-stop                Stop local test containers")
	fmt.Fprintln(w, "  env-migrate             Apply migrations to the test database")
	fmt.Fprintln(w, "  env-test                Run the local integration test flow")
}

func (c *CLI) runDBNew(args []string) error {
	fs := flag.NewFlagSet("db-new", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	name := fs.String("name", "", "migration name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return errors.New("db-new requires --name")
	}

	return c.DBNew(*name)
}

func (c *CLI) runDBRollback(args []string) error {
	fs := flag.NewFlagSet("db-rollback", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	count := fs.Int("count", 1, "number of migrations to roll back")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *count < 1 {
		return errors.New("db-rollback requires --count >= 1")
	}

	return c.DBRollback(*count)
}

func (c *CLI) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *CLI) runShell(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *CLI) dbURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		getEnv("DB_USER", defaultDBUser),
		getEnv("DB_PASSWORD", defaultDBPassword),
		getEnv("DB_HOST", defaultDBHost),
		getEnvInt("DB_PORT", defaultDBPort),
		getEnv("DB_NAME", defaultDBName),
		getEnv("DB_SSL_MODE", defaultDBSSLMode),
	)
}

func (c *CLI) migrationDir() string {
	return "file://db/migrations"
}

func (c *CLI) schemaURL() string {
	return "file://db/schema.sql"
}

func (c *CLI) devDBURL() string {
	return testutils.DevDBURL
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
