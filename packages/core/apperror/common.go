package apperror

// Codes
const (
	CodeInternal        = "INTERNAL_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeValidation      = "VALIDATION_FAILED"
	CodeUnAuthenticated = "UNAUTHENTICATED"
	CodeForbidden       = "FORBIDDEN"
	CodeThirdParty      = "EXTERNAL_SERVICE_ERROR"
)

// Common Errors
var (
	ErrInternal = &AppError{
		Type:    TypeInternal,
		Code:    CodeInternal,
		Message: "Internal Server Error",
	}
	ErrForbidden = &AppError{
		Type:    TypeForbidden,
		Code:    CodeForbidden,
		Message: "Forbidden",
	}
	ErrNotFound = &AppError{
		Type:    TypeNotFound,
		Code:    CodeNotFound,
		Message: "Not Found",
	}
	ErrUnAuthenticated = &AppError{
		Type:    TypeUnAuthenticated,
		Code:    CodeUnAuthenticated,
		Message: "Unauthenticated",
	}
	ErrThirdParty = &AppError{
		Type:    TypeThirdParty,
		Code:    CodeThirdParty,
		Message: "External Service Error",
	}
	ErrValidation = &AppError{
		Type:    TypeValidation,
		Code:    CodeValidation,
		Message: "Validation Failed",
	}
)
