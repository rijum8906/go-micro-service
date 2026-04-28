package dbcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var useTestDB bool

var execCommand = &cobra.Command{
	Use:   "exec",
	Short: "Execute operations to db directly without atlas",
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply SQL schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		content, err := os.ReadFile("db/schema.sql")
		if err != nil {
			return fmt.Errorf("read db/schema.sql: %w", err)
		}

		fmt.Printf("\n🗄️  Setting up database for %s\n", config.DBName)

		pool, err := utils.GetConnPool(useTestDB, config)
		if err != nil {
			return err
		}
		defer pool.Close()

		// execute sql
		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			return err
		}

		fmt.Println("\n✅  Database schema applied successfully")

		return nil
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all the tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		pool, err := utils.GetConnPool(useTestDB, config)
		if err != nil {
			return err
		}
		defer pool.Close()

		_, err = pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS public CASCADE;")
		if err != nil {
			return err
		}
		_, err = pool.Exec(context.Background(), "CREATE SCHEMA public;")
		if err != nil {
			return err
		}

		fmt.Println("\n✅  All tables cleaned successfully")

		return nil
	},
}

var truncateCmd = &cobra.Command{
	Use:   "truncate",
	Short: "Truncate tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		pool, err := utils.GetConnPool(useTestDB, config)
		if err != nil {
			return err
		}
		defer pool.Close()

		_, err = pool.Exec(context.Background(), "DROP SCHEMA public CASCADE;")
		if err != nil {
			return err
		}
		_, err = pool.Exec(context.Background(), "CREATE SCHEMA public;")
		if err != nil {
			return err
		}

		content, err := os.ReadFile("db/schema.sql")
		if err != nil {
			return fmt.Errorf("read db/schema.sql: %w", err)
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	execCommand.PersistentFlags().BoolVarP(&useTestDB, "test", "t", false, "Run migrations for test environment database")
	execCommand.AddCommand(applyCmd)
	execCommand.AddCommand(cleanCmd)
	execCommand.AddCommand(truncateCmd)

	DBCmd.AddCommand(execCommand)
}
