package main

import (
	"flag"

	"github.com/rijum8906/relay/services/user/cmd/test-cli/commands"
)

func main() {
	// Define flags
	flagStartAll := flag.Bool("start-all", false, "Run tests")
	flagStopAll := flag.Bool("stop-all", false, "Stop tests")
	flagMigrate := flag.Bool("migrate", false, "Migrate database")
	flagRunTests := flag.Bool("run-tests", false, "Run tests")

	// Parse flags
	flag.Parse()

	cli := commands.NewCLI()

	if *flagStartAll {
		cli.StartAll()
	}

	if *flagStopAll {
		cli.StopAll()
	}

	if *flagMigrate {
		cli.Migrate()
	}

	if *flagRunTests {
		cli.RunTests()
	}
}
