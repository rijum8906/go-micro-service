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
func runGoTest(args ...string) error {
	baseArgs := []string{"test"}

	if verbose {
		baseArgs = append(baseArgs, "-v")
	}
	if runRegexp != "" {
		baseArgs = append(baseArgs, "-run", runRegexp)
	}

	finalArgs := append(baseArgs, args...)
	return utils.RunCommand("go", finalArgs...)
}

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all tests with race detection and coverage",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGoTest("./...", "-race", "-cover", "-coverprofile=coverage.out", "-covermode=atomic")
	},
}

var runUnitTestsCmd = &cobra.Command{
	Use:   "unit",
	Short: "Run unit tests (excludes integration tags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Assumes integration tests use // +build integration
		return runGoTest("./...", "-short", "-race")
	},
}

var runIntegrationTestsCmd = &cobra.Command{
	Use:   "integration",
	Short: "Run integration tests using build tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGoTest("./...", "-tags=integration", "-race", "-v")
	},
}

var runBenchmarkTestsCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run performance benchmarks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGoTest("./...", "-bench=.", "-benchmem", "-run=^$") // -run=^$ skips units, runs only bench
	},
}

var testCoverageCmd = &cobra.Command{
	Use:   "cover",
	Short: "Generate and view coverage report in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n📊 Generating coverage report")
		if err := utils.RunCommand("go", "test", "./...", "-coverprofile=coverage.out"); err != nil {
			return err
		}

		return utils.RunCommand("go", "tool", "cover", "-html=coverage.out")
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
