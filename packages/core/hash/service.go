package hash

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"

	"github.com/rijum8906/relay/packages/core/apperror"
	"golang.org/x/crypto/bcrypt"
)

// Hash generates a bcrypt hash from a password
func (s *HashService) Hash(password string) (string, *apperror.AppError) {
	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("failed to hash password").WithDetail("hash_error", err.Error())
	}

	return string(hashedBytes), nil
}

// Verify checks if a password matches its hash
func (s *HashService) Verify(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate generates a cryptographically secure random string of given size
func (s *HashService) Generate(size int) (string, *apperror.AppError) {
	// Calculate bytes needed for base64 encoding
	// Base64 uses 4 characters per 3 bytes, so we need (size * 3 + 3) / 4 bytes
	byteSize := (size * 3) / 4
	if (size*3)%4 != 0 {
		byteSize++
	}

	// Generate random bytes
	randomBytes := make([]byte, byteSize)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("failed to generate random string").WithDetail("err", err.Error())
	}

	// Encode to base64 and truncate to desired length
	encoded := base64.URLEncoding.EncodeToString(randomBytes)
	if len(encoded) > size {
		encoded = encoded[:size]
	}

	return encoded, nil
}

// GenerateBase32 generates a cryptographically secure random string of given size encoded in Base32 (no padding)
func (s *HashService) GenerateBase32(size int) (string, *apperror.AppError) {
	// Base32 uses 8 characters per 5 bytes, so we need (size * 5) / 8 bytes
	byteSize := (size * 5) / 8
	if (size*5)%8 != 0 {
		byteSize++
	}

	// Generate random bytes
	randomBytes := make([]byte, byteSize)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("failed to generate random string").WithDetail("err", err.Error())
	}

	// Use StdEncoding with NoPadding so it doesn't append "========" to the end
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	encoded := encoder.EncodeToString(randomBytes)

	// Truncate to desired length safely
	if len(encoded) > size {
		encoded = encoded[:size]
	}

	return encoded, nil
}
