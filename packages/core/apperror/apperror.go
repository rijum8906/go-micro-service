// Package apperror defines error types and functions for the application
package apperror

import (
	"fmt"
	"sync"
)

type Config struct {
	AppEnv string
}

var (
	config Config
	once   sync.Once
)

// SetConfig sets the configuration once (thread-safe)
func SetConfig(cfg Config) {
	once.Do(func() {
		config = cfg
	})
}

func GetConfig() Config {
	return config
}

type AppError struct {
	Code      ErrorCode
	Message   string
	Details   []Detail
	RequestID string
}

type Detail struct {
	Field   string
	Message string
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) WithMessage(message string) *AppError {
	e.Message = message
	return e
}

func (e *AppError) WithDetail(field, message string) *AppError {
	e.Details = append(e.Details, Detail{
		Field:   field,
		Message: message,
	})
	return e
}

func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

func (e *AppError) Error() string {
	if config.AppEnv != "production" {
		fmt.Printf("Error: %s : %v", e.Message, e.Details)
	}

	return fmt.Sprintf("[%s] %s (ID: %s)", e.Code, e.Message, e.RequestID)
}
