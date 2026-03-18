// Package errors contains standard API error structures.
package errors

type AppError struct {
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	Errors   []Error `json:"errors,omitempty"`
	Internal error   `json:"-"`
}

type Error struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return e.Internal.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// WithInternal returns a new AppError with the internal error set
func (e *AppError) WithInternal(err error) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Internal: err,
		Errors:   e.Errors,
	}
}

// WithField adds a specific validation error
func (e *AppError) WithField(field, message string) *AppError {
	newErrors := append(e.Errors, Error{Field: field, Message: message})
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Internal: e.Internal,
		Errors:   newErrors,
	}
}

func NewAppError(code string, message string, errors []Error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}
