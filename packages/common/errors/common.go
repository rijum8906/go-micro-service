package errors

var (
	// 400 Bad Request
	ErrBadRequest   = NewAppError("BAD_REQUEST", "bad request", nil)
	ErrInvalidBody  = NewAppError("INVALID_BODY", "invalid request body", nil)
	ErrInvalidInput = NewAppError("INVALID_INPUT", "invalid input provided", nil)
	ErrMissingField = NewAppError("MISSING_FIELD", "required field missing", nil)

	// 401 Unauthorized
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "unauthorized", nil)
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "invalid email or password", nil)
	ErrInvalidToken       = NewAppError("INVALID_TOKEN", "invalid or malformed token", nil)
	ErrInvalidTokenFormat = NewAppError("INVALID_TOKEN_FORMAT", "invalid token format", nil)
	ErrExpiredToken       = NewAppError("EXPIRED_TOKEN", "token has expired", nil)
	ErrInvalidTokenClaims = NewAppError("INVALID_TOKEN_CLAIMS", "invalid token claims", nil)
	ErrInvalidPassword    = NewAppError("INVALID_PASSWORD", "invalid password", nil)
	ErrInvalidEmail       = NewAppError("INVALID_EMAIL", "invalid email address", nil)
	ErrInvalidUser        = NewAppError("INVALID_USER", "user not found", nil)
	ErrInvalidScope       = NewAppError("INVALID_SCOPE", "invalid token scope", nil)

	// 403 Forbidden
	ErrForbidden   = NewAppError("FORBIDDEN", "access forbidden", nil)
	ErrInvalidRole = NewAppError("INVALID_ROLE", "insufficient permissions", nil)

	// 404 Not Found
	ErrNotFound   = NewAppError("NOT_FOUND", "resource not found", nil)
	ErrDBNotFound = NewAppError("DB_NOT_FOUND", "database record not found", nil)

	// 409 Conflict
	ErrConflict   = NewAppError("CONFLICT", "resource conflict", nil)
	ErrUserExists = NewAppError("USER_EXISTS", "user already exists", nil)
	ErrDBConflict = NewAppError("DB_CONFLICT", "database conflict", nil)

	// 422 Validation
	ErrValidation = NewAppError("VALIDATION_ERROR", "validation failed", nil)

	// Database Errors
	ErrDBConnection = NewAppError("DB_CONNECTION_ERROR", "database connection error", nil)
	ErrDBError      = NewAppError("DB_ERROR", "database error", nil)

	// Internal Server Errors
	ErrInternal = NewAppError("INTERNAL_SERVER_ERROR", "internal server error", nil)
)
