package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// Config contains only the fields required to boot the task-service skeleton.
type Config struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	AppName string `env:"APP_NAME" envDefault:"task-service"`
	Port    int    `env:"PORT" envDefault:"8083"`

	LogLevel     string `env:"LOG_LEVEL" envDefault:"debug"`
	EnableJSON   bool   `env:"ENABLE_JSON" envDefault:"false"`
	EnableCaller bool   `env:"ENABLE_CALLER" envDefault:"false"`
	EnableStack  bool   `env:"ENABLE_STACK" envDefault:"false"`
	LogFile      string `env:"LOG_FILE"`
}

// Load reads .env values for local development and parses the task-service config.
func Load() (*Config, *apperror.AppError) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, apperror.ErrInternal.WithMessage(
			fmt.Sprintf("failed to parse task-service environment variables: %v", err),
		)
	}

	return cfg, nil
}

// MustLoad is a helper for main packages that should fail fast on invalid config.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("critical task-service config error: %v", err)
	}

	return cfg
}
