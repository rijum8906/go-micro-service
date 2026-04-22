// Package dbcmd
package dbcmd

import (
	"fmt"

	"github.com/rijum8906/relay/rcli/utils"
	"github.com/spf13/cobra"
)

var DBCMd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Setting up database...")
		pool := utils.MustConnectDB(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, "postgres", config.DBSSLMode)
		err := utils.CreateDatabase(pool, "dev_"+config.DBName)
		if err != nil {
			panic(err)
		}
		err = utils.CreateDatabase(pool, config.DBName)
		if err != nil {
			panic(err)
		}
		fmt.Println("✅ Database initialized")
	},
}

func init() {
	DBCMd.AddCommand(initCmd)
}
