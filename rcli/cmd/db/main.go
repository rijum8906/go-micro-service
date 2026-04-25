// Package dbcmd
package dbcmd

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var DBCmd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadEnv()
		if err != nil {
			return err
		}

		pool, err := utils.ConnectDB(config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		if err != nil {
			return err
		}
		defer pool.Close()

		var exists bool
		err = pool.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
			config.DBName).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			// Create database
			sql := fmt.Sprintf("CREATE DATABASE %s OWNER = %s ENCODING = 'UTF8'",
				config.DBName, config.DBUser)
			_, err = pool.Exec(context.Background(), sql)
			return err
		}

		fmt.Println("\n✅  Database initialized successfully")

		return nil
	},
}

func init() {
	DBCmd.AddCommand(initCmd)
}
