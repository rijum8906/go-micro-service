// Package dbcmd
package dbcmd

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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

		pool := &pgxpool.Pool{}

		if useTestDB {
			fmt.Println("\n🧪 Setting up test database")
			pool, err = utils.ConnectDB(utils.WithDBName("test_db"),
				utils.WithPort(5433),
				utils.WithUser("test_user"),
				utils.WithPassword("test_password"))
			if err != nil {
				return err
			}

			fmt.Println("\n🧪 Creating database" + utils.GetTestDBName(config.AppName))
			if err = utils.CreateDatabase(pool, utils.GetTestDBName(config.AppName), utils.WithPort(5433)); err != nil {
				return err
			}
		} else {
			pool, err = utils.ConnectDB()
			if err != nil {
				return err
			}

			fmt.Println("\n🧪 Creating database" + config.DBName)
			if err := utils.CreateDatabase(pool, config.DBName); err != nil {
				return err
			}
		}

		defer pool.Close()

		fmt.Println("\n✅  Database initialized successfully")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&useTestDB, "test", "t", false, "Run migrations for test environment database")
	DBCmd.AddCommand(initCmd)
}
