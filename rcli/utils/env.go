package utils

import (
	"fmt"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Environment struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	AppName string `env:"APP_NAME,required"`

	// Postgres
	DBHost     string `env:"DB_HOST,required"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
}

func LoadEnv() (*Environment, error) {
	_ = godotenv.Load()

	cfg := &Environment{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("load environment: %w", err)
	}

	return cfg, nil
}
