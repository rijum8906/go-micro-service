package testutils

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/db"
)

var DBConf = db.Config{
	Host:     "localhost",
	Port:     5433,
	User:     "test_user",
	Password: "test_password",
	DBName:   "test_db",
	SSLMode:  "disable",
}

func MustConnectDB() *pgxpool.Pool {
	pool, appErr := db.Connect(context.Background(), DBConf)
	if appErr != nil {
		panic(appErr)
	}

	return pool
}
