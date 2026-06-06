// Package apperror defines error types and functions for the application
package apperror

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Frame struct {
	File string
	Line int
	Func string
}

type Config struct {
	Logger *zap.Logger
	AppEnv string
	Debug  bool
}

var (
	config *Config
	once   sync.Once
)

// SetConfig sets the configuration once (thread-safe)
func SetConfig(cfg *Config) {
	once.Do(func() {
		config = cfg
	})
}

func GetConfig() *Config {
	return config
}

type AppError struct {
	Code      ErrorCode
	Message   string
	Details   []Detail
	RequestID string
	Frames    []Frame // Captured stack frames
	isLogged  bool
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

// NewWithFrame captures the caller's line number
func NewWithFrame(code ErrorCode, message string, skip int) *AppError {
	frames := captureFrames(skip)

	return &AppError{
		Code:    code,
		Message: message,
		Frames:  frames,
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

func (e *AppError) WithFrames(frames []Frame) *AppError {
	e.Frames = frames
	return e
}

func (e *AppError) Log() {
	if !e.isLogged {
		if config.Logger == nil {
			fmt.Printf("[%s] %s Details : %v Frames : %v \n", e.Code, e.Message, e.Details, e.Frames)
		} else {
			config.Logger.Error(e.Message, ParseAppErrorIntoZapFields(e)...)
		}

		e.isLogged = true
	}
}

func (e *AppError) Error() string {
	if e.Code == CodeInternal || e.Code == CodeThirdParty {
		e.Log()
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)

}
