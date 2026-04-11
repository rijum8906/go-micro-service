package testutils

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/db"
)

const (
	DBHost     = "localhost"
	DBPort     = 5433
	DBUser     = "test_user"
	DBPassword = "test_password"
	DBName     = "test_db"
	DBSSLMode  = "disable"
	DevDBURL   = "docker://postgres/17/dev?search_path=public"
)

var DBConf = db.Config{
	Host:        DBHost,
	Port:        DBPort,
	User:        DBUser,
	Password:    DBPassword,
	DBName:      DBName,
	SSLMode:     DBSSLMode,
	RetryCounts: 5,
}

func MustConnectDB() *pgxpool.Pool {
	pool, appErr := db.Connect(context.Background(), DBConf)
	if appErr != nil {
		panic(appErr)
	}

	return pool
}
