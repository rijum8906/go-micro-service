package apperror

type ErrorCode string

const (
	// Core
	CodeInternal         ErrorCode = "INTERNAL"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeValidation       ErrorCode = "VALIDATION"
	CodeUnAuthenticated  ErrorCode = "UNAUTHENTICATED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeBadRequest       ErrorCode = "BAD_REQUEST"
	CodeTimeout          ErrorCode = "TIMEOUT"
	CodeRateLimit        ErrorCode = "RATE_LIMIT"
	CodeThirdParty       ErrorCode = "THIRD_PARTY"
	CodeDatabase         ErrorCode = "DATABASE_ERROR"
	CodeCache            ErrorCode = "CACHE_ERROR"
	CodeDependency       ErrorCode = "DEPENDENCY_ERROR"
	CodePrecondition     ErrorCode = "PRECONDITION_FAILED"
	CodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"
	CodeUnavailable      ErrorCode = "SERVICE_UNAVAILABLE"

	// Token
	CodeTokenInvalidSignature ErrorCode = "INVALID_SIGNATURE"
	CodeTokenExpired          ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid          ErrorCode = "INVALID_TOKEN"
	CodeTokenMalformed        ErrorCode = "TOKEN_MALFORMED"
)

// Common Errors
var (
	ErrInternal   = NewWithFrame(CodeInternal, "Internal Server Error", 1)
	ErrThirdParty = NewWithFrame(CodeThirdParty, "Third Party Service Error", 1)

	ErrForbidden = &AppError{
		Code:    CodeForbidden,
		Message: "Forbidden",
	}
	ErrPermissionDenied = &AppError{
		Code:    CodePermissionDenied,
		Message: "Permission Denied",
	}
	ErrNotFound = &AppError{
		Code:    CodeNotFound,
		Message: "Not Found",
	}
	ErrUnAuthenticated = &AppError{
		Code:    CodeUnAuthenticated,
		Message: "Unauthenticated",
	}
	ErrValidation = &AppError{
		Code:    CodeValidation,
		Message: "Validation Failed",
	}
	ErrConflict = &AppError{
		Code:    CodeConflict,
		Message: "Conflict",
	}
)
