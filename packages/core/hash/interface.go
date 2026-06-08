// Package hash provides hashing functions
package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type HashService struct {
	cost int
}

type Config struct {
	BcryptCost int
}

func NewHashService(cfg Config) *HashService {
	cost := cfg.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost // 10
	}

	// Validate cost range
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &HashService{
		cost: cost,
	}
}
