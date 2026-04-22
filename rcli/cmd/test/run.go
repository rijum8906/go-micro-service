package testcmd

import (
	"fmt"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var (
	verbose   bool
	runRegexp string
)

// Helper to centralize test execution logic
func runGoTest(args ...string) {
	baseArgs := []string{"test"}

	if verbose {
		baseArgs = append(baseArgs, "-v")
	}
	if runRegexp != "" {
		baseArgs = append(baseArgs, "-run", runRegexp)
	}

	finalArgs := append(baseArgs, args...)
	utils.RunCommand("go", finalArgs...)
}

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all tests with race detection and coverage",
	Run: func(cmd *cobra.Command, args []string) {
		runGoTest("./...", "-race", "-cover", "-coverprofile=coverage.out", "-covermode=atomic")
	},
}

var runUnitTestsCmd = &cobra.Command{
	Use:   "unit",
	Short: "Run unit tests (excludes integration tags)",
	Run: func(cmd *cobra.Command, args []string) {
		// Assumes integration tests use // +build integration
		runGoTest("./...", "-short", "-race")
	},
}

var runIntegrationTestsCmd = &cobra.Command{
	Use:   "integration",
	Short: "Run integration tests using build tags",
	Run: func(cmd *cobra.Command, args []string) {
		runGoTest("./...", "-tags=integration", "-race", "-v")
	},
}

var runBenchmarkTestsCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run performance benchmarks",
	Run: func(cmd *cobra.Command, args []string) {
		runGoTest("./...", "-bench=.", "-benchmem", "-run=^$") // -run=^$ skips units, runs only bench
	},
}

var testCoverageCmd = &cobra.Command{
	Use:   "cover",
	Short: "Generate and view coverage report in browser",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating coverage report...")
		utils.RunCommand("go", "test", "./...", "-coverprofile=coverage.out")
		utils.RunCommand("go", "tool", "cover", "-html=coverage.out")
	},
}

func init() {
	// Persistent flags available to all subcommands
	TestCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	TestCmd.PersistentFlags().StringVarP(&runRegexp, "run", "r", "", "run only tests matching regexp")

	// Compose command tree
	TestCmd.AddCommand(testRunCmd)
	TestCmd.AddCommand(testCoverageCmd)

	testRunCmd.AddCommand(runUnitTestsCmd)
	testRunCmd.AddCommand(runIntegrationTestsCmd)
	testRunCmd.AddCommand(runBenchmarkTestsCmd)
}
