package main

import (
	"flag"

	"github.com/rijum8906/relay/cmd/commands"
)

func main() {
	// define flags
	flagSetup := flag.Bool("setup", false, "setup the project")
	flagDev := flag.Bool("dev", false, "run docker compose")

	flag.Parse()

	if *flagSetup {
		commands.Setup()
	}

	if *flagDev {
		commands.RunDockerCompose()
	}
}
