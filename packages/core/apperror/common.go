package apperror

type ErrorCode string

const (
	// Core
	CodeInternal        ErrorCode = "INTERNAL"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeValidation      ErrorCode = "VALIDATION"
	CodeUnAuthenticated ErrorCode = "UNAUTHENTICATED"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeConflict        ErrorCode = "CONFLICT"
	CodeBadRequest      ErrorCode = "BAD_REQUEST"
	CodeTimeout         ErrorCode = "TIMEOUT"
	CodeRateLimit       ErrorCode = "RATE_LIMIT"
	CodeThirdParty      ErrorCode = "THIRD_PARTY"
	CodeDatabase        ErrorCode = "DATABASE_ERROR"
	CodeCache           ErrorCode = "CACHE_ERROR"
	CodeDependency      ErrorCode = "DEPENDENCY_ERROR"
	CodePrecondition    ErrorCode = "PRECONDITION_FAILED"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	CodeUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"

	// Token
	CodeTokenInvalidSignature ErrorCode = "INVALID_SIGNATURE"
	CodeTokenExpired          ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid          ErrorCode = "INVALID_TOKEN"
)

// Common Errors
var (
	ErrInternal = &AppError{
		Code:    CodeInternal,
		Message: "Internal Server Error",
	}
	ErrForbidden = &AppError{
		Code:    CodeForbidden,
		Message: "Forbidden",
	}
	ErrNotFound = &AppError{
		Code:    CodeNotFound,
		Message: "Not Found",
	}
	ErrUnAuthenticated = &AppError{
		Code:    CodeUnAuthenticated,
		Message: "Unauthenticated",
	}
	ErrThirdParty = &AppError{
		Code:    CodeThirdParty,
		Message: "External Service Error",
	}
	ErrValidation = &AppError{
		Code:    CodeValidation,
		Message: "Validation Failed",
	}
)
