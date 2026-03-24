// Package hash provides hashing functions
package hash

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"golang.org/x/crypto/bcrypt"
)

type HashService interface {
	Hash(password string) (string, *apperror.AppError)
	Verify(hash string, password string) *apperror.AppError
	Generate(size int) (string, *apperror.AppError)
}

type hashService struct {
	cost int
}

type Config struct {
	BcryptCost int
}

func NewHashService(cfg Config) *hashService {
	cost := cfg.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost // 10
	}

	// Validate cost range
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &hashService{
		cost: cost,
	}
}
