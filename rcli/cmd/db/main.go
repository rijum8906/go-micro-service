// Package dbcmd
package dbcmd

import (
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

		if err = utils.CreateDatabase(pool, config.DBName); err != nil {
			return err
		}
		if err = utils.CreateDatabase(pool, utils.DevDBName); err != nil {
			return err
		}

		fmt.Println("\n✅  Database initialized successfully")

		return nil
	},
}

func init() {
	DBCmd.AddCommand(initCmd)
}
