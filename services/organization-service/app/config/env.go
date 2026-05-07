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

	// TTL
	InvitationTokenTTL int `env:"INVITATION_TOKEN_TTL" envDefault:"7"` // in days

	// Used for testing
	FGATestAPIURL      string `env:"FGA_TEST_API_URL" envDefault:"localhost:9000"`
	FGATestStoreID     string `env:"FGA_TEST_STORE_ID" envDefault:""`
	FGATestAuthModelID string `env:"FGA_TEST_AUTH_MODEL_ID" envDefault:""`
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
