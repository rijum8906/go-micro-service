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
