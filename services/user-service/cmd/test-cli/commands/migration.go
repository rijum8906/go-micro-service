package commands

import (
	"fmt"
	"os/exec"

	"github.com/rijum8906/relay/packages/core/testutils"
)

func (c *CLI) Migrate() {
	println("Migrating database...")
	DB_URL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=public",
		testutils.DBUser, testutils.DBPassword, testutils.DBHost, testutils.DBPort, testutils.DBName, testutils.DBSSLMode)

	cmd := exec.Command(
		"atlas", "migrate", "apply",
		"--url", DB_URL,
		"--dir", "file://./db/migrations",
	)

	err := cmd.Run()
	if err != nil {
		println("Migration failed")
		panic(err)
	}
	println("Migration successful")
}
