package errors

type AppError struct {
	StatusCode int      `json:"-"`
	Message    string   `json:"message"`
	Errors     *[]Error `json:"errors,omitempty"`
}

type Error struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string, errors *[]Error) *AppError {
	return &AppError{
		StatusCode: code,
		Message:    message,
		Errors:     errors,
	}
}
