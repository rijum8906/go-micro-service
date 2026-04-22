package testcmd

import (
	"github.com/rijum8906/relay/cli/utils"
	"github.com/spf13/cobra"
)

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("go", "test", "./...", "-race", "-cover", "-coverprofile=coverage.out", "-covermode=atomic")
	},
}

var runUnitTestsCmd = &cobra.Command{
	Use:   "unit",
	Short: "Run unit tests",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("go", "test", "./...", "-race", "-cover", "-coverprofile=coverage.out", "-covermode=atomic")
	},
}

var runIngerateTestsCmd = &cobra.Command{
	Use:   "integration",
	Short: "Run integration tests",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("go", "test", "./...", "-race", "-cover", "-coverprofile=coverage.out", "-covermode=atomic")
	},
}

var runBenchmarkTestsCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run benchmark tests",
	Run: func(cmd *cobra.Command, args []string) {
		utils.RunCommand("go", "test", "./...", "-bench=.", "-benchmem", "-benchtime=5s")
	},
}

func init() {
	testRunCmd.AddCommand(runUnitTestsCmd)
	testRunCmd.AddCommand(runIngerateTestsCmd)
	testRunCmd.AddCommand(runBenchmarkTestsCmd)
}
