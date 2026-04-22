// Package dbcmd
package dbcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	Run: func(cmd *cobra.Command, args []string) {
		// Find all schema.sql files
		services, err := filepath.Glob("services/*/db/schema.sql")
		if err != nil {
			fmt.Printf("❌ Failed to find services: %v\n", err)
			return
		}

		if len(services) == 0 {
			fmt.Println("⚠️  No services found with db/schema.sql")
			return
		}

		fmt.Printf("🔧 Setting up database for %d service(s)...\n", len(services))

		var failedServices []string

		for _, schemaPath := range services {
			// Extract service directory
			// schemaPath: services/auth/db/schema.sql
			// serviceDir: services/auth
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

			fmt.Printf("✅ Service %s initialized\n", serviceName)
		}

		// Summary
		fmt.Println("\n" + strings.Repeat("=", 50))
		if len(failedServices) > 0 {
			fmt.Printf("❌ Failed services: %s\n", strings.Join(failedServices, ", "))
			fmt.Println("⚠️  Database initialization completed with errors")
		} else {
			fmt.Println("✅ All databases initialized successfully")
		}
	},
}

func init() {
	initCmd.Flags().BoolVarP(&useAtlas, "atlas", "a", false, "use Atlas database")
	DBCMd.AddCommand(initCmd)
}
