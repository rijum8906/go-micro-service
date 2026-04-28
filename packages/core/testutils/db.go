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

type DBConfig = db.Config

type Option func(*DBConfig)

func WithHost(host string) Option {
	return func(c *DBConfig) {
		c.Host = host
	}
}

func WithPort(count int) Option {
	return func(c *DBConfig) {
		c.Port = count
	}
}

func WithUser(user string) Option {
	return func(c *DBConfig) {
		c.User = user
	}
}

func WithPassword(password string) Option {
	return func(c *DBConfig) {
		c.Password = password
	}
}

func WithDBName(name string) Option {
	return func(c *DBConfig) {
		c.DBName = name
	}
}

func WithSSLMode(mode string) Option {
	return func(c *DBConfig) {
		c.SSLMode = mode
	}
}

func MustConnectDB(options ...Option) *pgxpool.Pool {
	config := DBConfig{
		Host:        DBHost,
		Port:        DBPort,
		User:        DBUser,
		Password:    DBPassword,
		DBName:      DBName,
		SSLMode:     DBSSLMode,
		RetryCounts: 5,
	}

	// Apply options
	for _, opt := range options {
		opt(&config)
	}

	pool, appErr := db.Connect(context.Background(), config)
	if appErr != nil {
		panic(appErr)
	}

	return pool
}
