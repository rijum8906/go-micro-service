// Package env provides environment configuration
package env

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	AppName string `env:"APP_NAME,required"`
	Port    int    `env:"PORT" envDefault:"8080"`

	// Postgres
	DBHost     string `env:"DB_HOST,required"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASS,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`

	// Redis
	RedisHost string `env:"REDIS_HOST,required"`
	RedisPort int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPass string `env:"REDIS_PASS"`

	// JWT
	JWTSecret    string `env:"JWT_SECRET,required"`
	ScopedSecret string `env:"SCOPED_SECRET,required"` // For scoped tokens

	// TTL
	SessionTTL      time.Duration `env:"SESSION_TTL" envDefault:"15m"`
	ScopedTokenTTL  time.Duration `env:"SCOPED_TOKEN_TTL" envDefault:"10m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"30d"`
}

// Load reads .env files (for local dev) and parses environment variables
func Load() (*Config, error) {
	// 1. Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{}

	// 2. Parse variables into the struct
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	return cfg, nil
}

// MustLoad is a helper for main.go that panics if config is missing
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Critical Config Error: %v", err)
	}
	return cfg
}
