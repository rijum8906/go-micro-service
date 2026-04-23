// Package dbcmd
package dbcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var useAtlas bool

var DBCMd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsRootDir() {
			return fmt.Errorf("must be run from the project root directory")
		}
		// Find all schema.sql files
		services, err := filepath.Glob("services/*/db/schema.sql")
		if err != nil {
			return fmt.Errorf("find service schemas: %w", err)
		}

		if len(services) == 0 {
			fmt.Println("\n⚠️  No services found with db/schema.sql")
			return nil
		}

		fmt.Printf("\n🗄️  Setting up databases for %d service(s)\n", len(services))

		var failedServices []string

		for _, schemaPath := range services {
			serviceDir := filepath.Dir(filepath.Dir(schemaPath))
			serviceName := filepath.Base(serviceDir)

			fmt.Printf("\n📦 Processing service: %s\n", serviceName)

			// Prepare command
			var cmdArgs []string
			if useAtlas {
				cmdArgs = []string{"db", "migrate", "apply"}
			} else {
				cmdArgs = []string{"db", "migrate", "sql-apply"}
			}

			// Execute command in service directory
			c := exec.Command("rcli", cmdArgs...)
			c.Dir = serviceDir
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				fmt.Printf("❌ Failed to initialize %s: %v\n", serviceName, err)
				failedServices = append(failedServices, serviceName)
				continue
			}

			fmt.Printf("✅ Initialized service: %s\n", serviceName)
		}

		// Summary
		fmt.Printf("\n%s\n", strings.Repeat("=", 50))
		if len(failedServices) > 0 {
			return fmt.Errorf("database initialization failed for: %s", strings.Join(failedServices, ", "))
		} else {
			fmt.Println("✅ All databases initialized successfully")
		}

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&useAtlas, "atlas", "a", false, "use Atlas database")
	DBCMd.AddCommand(initCmd)
}
