package main

import (
	"flag"

	"github.com/rijum8906/relay/cmd/commands"
)

func main() {
	// define flags
	flagSetup := flag.Bool("setup", false, "setup the project")

	flag.Parse()

	if *flagSetup {
		commands.Setup()
	}
}
