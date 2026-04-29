// Package utils
package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DevDBName = "dev_db"

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Option func(*Config)

func WithPort(count int) Option {
	return func(c *Config) {
		c.Port = count
	}
}

func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

func WithDBName(name string) Option {
	return func(c *Config) {
		c.DBName = name
	}
}

func WithUser(user string) Option {
	return func(c *Config) {
		c.User = user
	}
}

// ConnectDB connects to the database and verifies that it is reachable.
func ConnectDB(options ...Option) (*pgxpool.Pool, error) {
	defaultOption := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "postgres",
		SSLMode:  "disable",
	}

	// Apply options
	for _, opt := range options {
		opt(&defaultOption)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		defaultOption.User,
		defaultOption.Password,
		defaultOption.Host,
		defaultOption.Port,
		defaultOption.DBName,
		defaultOption.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	// Test connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database is unreachable: %w", err)
	}

	return pool, nil
}

func GetTestDBName(serviceName string) string {
	serviceName = strings.Split(serviceName, "-")[0]
	serviceName = strings.Split(serviceName, "_")[0]
	serviceName = strings.Split(serviceName, " ")[0]
	return fmt.Sprintf("test_%s", strings.ToLower(serviceName))
}

// GetConnPool returns the database connection pool
// if useTestDB is true, it will connect to the test database
// if useTestDB is false, it will connect to the production database
func GetConnPool(useTestDB bool, config *Environment) (*pgxpool.Pool, error) {
	if useTestDB {
		return ConnectDB(WithDBName(
			GetTestDBName(config.AppName)),
			WithUser("test_user"),
			WithPassword("test_password"),
			WithPort(5433))
	}

	return ConnectDB(
		WithDBName(config.DBName),
		WithUser(config.DBUser),
		WithPassword(config.DBPassword),
		WithPort(config.DBPort))
}

func CreateDatabase(pool *pgxpool.Pool, name string) error {
	ctx := context.Background()

	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		name,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check whether database %q exists: %w", name, err)
	}

	if exists {
		return nil // Already exists
	}

	// Create database (can't use parameters for database names)
	sql := "CREATE DATABASE " + name
	_, err = pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("create database %q: %w", name, err)
	}

	return nil
}
