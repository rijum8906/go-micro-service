package dto_test

import (
	"testing"

	validator "github.com/go-playground/validator/v10"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/stretchr/testify/assert"
)

var validate = validator.New()

func TestSignUpDTO(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := dto.SignUpDTO{
			Email:     "zPjgP@example.com",
			Password:  "password",
			FirstName: "John",
			LastName:  "Doe",
			Metadata: dto.UserMetadata{
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
				IPAddr:    "127.0.0.1",
				DeviceID:  "1234567890",
			},
		}

		err := validate.Struct(input)
		assert.NoError(t, err)
	})

	t.Run("required", func(t *testing.T) {
		input := dto.SignUpDTO{
			Email:    "zPjgP@example.com",
			Password: "passdskhfkhfkhahs",
			// FirstName: "John",
			LastName: "Doe",
			Metadata: dto.UserMetadata{
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
				IPAddr:    "127.0.0.1",
				DeviceID:  "1234567890",
			},
		}

		err := validate.Struct(input)
		assert.Error(t, err)
	})

	t.Run("inavlid length", func(t *testing.T) {
		input := dto.SignUpDTO{
			Email:     "zPjgP@example.com",
			Password:  "pass",
			FirstName: "John",
			LastName:  "Doe",
			Metadata: dto.UserMetadata{
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
				IPAddr:    "127.0.0.1",
				DeviceID:  "1234567890",
			},
		}

		err := validate.Struct(input)
		assert.Error(t, err)
	})
}

func TestSignInDTO(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := dto.SignInDTO{
			Email:    "zPjgP@example.com",
			Password: "password",
			Metadata: dto.UserMetadata{
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
				IPAddr:    "127.0.0.1",
				DeviceID:  "1234567890",
			},
		}

		err := validate.Struct(input)
		assert.NoError(t, err)
	})

	t.Run("required", func(t *testing.T) {
		input := dto.SignInDTO{
			Email:    "zPjgP@example.com",
			Password: "passdskhfkhfkhahs",
		}

		err := validate.Struct(input)
		assert.Error(t, err)
	})
}
