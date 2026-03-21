// Package apperror defines error types and functions for the application
package apperror

import "fmt"

type ErrorType int

const (
	TypeValidation ErrorType = iota
	TypeUnAuthenticated
	TypeForbidden
	TypeNotFound
	TypeInternal
	TypeThirdParty
	TypeUnknown
)

type AppError struct {
	Type      ErrorType
	Code      string
	Message   string
	Details   []Detail
	RequestID string
}

type Detail struct {
	Field   string
	Message string
}

func New(errType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errType,
		Code:    code,
		Message: message,
	}
}

func (e *AppError) AddDetail(filed, message string) {
	e.Details = append(e.Details, Detail{
		Field:   filed,
		Message: message,
	})
}

func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s (ID: %s)", e.Code, e.Message, e.RequestID)
}
