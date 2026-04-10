package main

import (
	"fmt"
	"os"

	"github.com/rijum8906/relay/services/user/cmd/test-cli/commands"
)

func main() {
	cli := commands.NewCLI()
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
