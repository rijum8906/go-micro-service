// Package config
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreenv"
)

type Env struct {
	coreenv.CoreEnv
	FGAAPIURL string `env:"FGA_API_URL" envDefault:"localhost:8000"`

	// SMTP
	SMTPHost        string `env:"SMTP_HOST,required"`
	SMTPPort        int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername    string `env:"SMTP_USERNAME,required"`
	SMTPPassword    string `env:"SMTP_PASSWORD,required"`
	SMTPFromEmail   string `env:"SMTP_FROM_EMAIL,required"`
	SMTPFromName    string `env:"SMTP_FROM_NAME,required"`
	SMTPUseTLS      bool   `env:"SMTP_USE_TLS" envDefault:"true"`
	SMTPUseStartTLS bool   `env:"SMTP_USE_STARTTLS" envDefault:"false"`
	SMTPTimeout     int    `env:"SMTP_TIMEOUT" envDefault:"10"`
	SMTPRetries     int    `env:"SMTP_RETRIES" envDefault:"3"`
}

func LoadEnv() (*Env, *apperror.AppError) {
	// 1. Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Env{}

	// 2. Parse variables into the struct
	if err := env.Parse(cfg); err != nil {
		return nil, apperror.ErrInternal.WithMessage(fmt.Sprintf("Failed to parse environment variables: %v", err))
	}

	return cfg, nil
}
