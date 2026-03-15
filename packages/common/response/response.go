// Package response contains standard API response structures.
package response

import "time"

// ============================================================================
// Base Structure
// ============================================================================

// Base contains common fields for all responses
type Base struct {
	Success   bool      `json:"success" example:"true"`
	RequestID string    `json:"requestId,omitempty" example:"req-123"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// NewBase creates a new base response
func NewBase(success bool) Base {
	return Base{
		Success:   success,
		Timestamp: time.Now().UTC(),
	}
}

// WithRequestID adds a request ID to the base
func (b Base) WithRequestID(requestID string) Base {
	b.RequestID = requestID
	return b
}

// ============================================================================
// Success Response
// ============================================================================

// Success represents a successful API response
type Success[T any] struct {
	Base
	Message string `json:"message,omitempty" example:"User created successfully"`
	Data    T      `json:"data,omitempty"`
}

// NewSuccess creates a new success response
func NewSuccess[T any](data T) *Success[T] {
	return &Success[T]{
		Base: NewBase(true),
		Data: data,
	}
}

// NewSuccessWithMessage creates a success response with a message
func NewSuccessWithMessage[T any](data T, message string) *Success[T] {
	return &Success[T]{
		Base:    NewBase(true),
		Message: message,
		Data:    data,
	}
}

// WithRequestID adds a request ID to the success response
func (s *Success[T]) WithRequestID(requestID string) *Success[T] {
	s.Base = s.Base.WithRequestID(requestID)
	return s
}

// WithMessage adds a message to the success response
func (s *Success[T]) WithMessage(message string) *Success[T] {
	s.Message = message
	return s
}

// ============================================================================
// Error Response
// ============================================================================

// Error represents a standardized error response
type Error struct {
	Base
	Code    string       `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Errors  []FieldError `json:"errors,omitempty"`
	Message string       `json:"message" example:"Invalid input"`
}

// FieldError represents a field-specific validation error
type FieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"must be a valid email address"`
}

// NewError creates a new error response
func NewError(message string, code string) *Error {
	return &Error{
		Base:    NewBase(false),
		Message: message,
		Code:    code,
	}
}

// WithRequestID adds a request ID to the error response
func (e *Error) WithRequestID(requestID string) *Error {
	e.Base = e.Base.WithRequestID(requestID)
	return e
}

// WithFieldErrors adds field-specific errors to the error response
func (e *Error) WithFieldErrors(errors []FieldError) *Error {
	e.Errors = errors
	return e
}

// WithCode adds or updates the error code
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// ============================================================================
// Helper Functions
// ============================================================================

// IsSuccess returns true if the response is a success response
func IsSuccess(success bool) bool {
	return success
}

// IsError returns true if the response is an error response
func IsError(success bool) bool {
	return !success
}
