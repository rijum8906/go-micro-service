package testutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/db"
)

// NOTE: test db naming convension
// - test_{service_first_name}
// - eg. test_user

// GetTestDBName returns the test database name based on service
func GetTestDBName(serviceName string) string {
	// test_user, test_organization, test_auth
	return fmt.Sprintf("test_%s", strings.ToLower(strings.TrimSpace(strings.Split(serviceName, " ")[0])))
}

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
		Host:        "localhost",
		Port:        5432,
		User:        "test_user",
		Password:    "test_password",
		DBName:      "test_db",
		SSLMode:     "disable",
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
