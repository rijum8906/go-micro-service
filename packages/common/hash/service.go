// Package hash
package hash

import "golang.org/x/crypto/bcrypt"

type Service interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

type service struct {
	cost int
}

func NewService(cost int) Service {
	return &service{
		cost: cost,
	}
}

func (s *service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *service) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
