package errors

var (
	// HTTP Responses
	ErrNotFound     = NewAppError(404, "not found", []Error{})
	ErrInternal     = NewAppError(500, "internal server error", []Error{})
	ErrUnauthorized = NewAppError(401, "unauthorized", []Error{})
	ErrForbidden    = NewAppError(403, "forbidden", []Error{})
	ErrBadRequest   = NewAppError(400, "bad request", []Error{})
	ErrConflict     = NewAppError(409, "conflict", []Error{})

	// Database errors
	ErrDBConnection = NewAppError(500, "database connection error", []Error{})
	ErrDBNotFound   = NewAppError(404, "database not found", []Error{})
	ErrDBConflict   = NewAppError(409, "database conflict", []Error{})

	// Validation errors
	ErrValidation         = NewAppError(400, "validation error", []Error{})
	ErrInvalidToken       = NewAppError(401, "invalid token", []Error{})
	ErrInvalidBody        = NewAppError(400, "invalid body", []Error{})
	ErrInvalidCredentials = NewAppError(401, "invalid credentials", []Error{})

	// Authentication errors
	ErrInvalidPassword = NewAppError(401, "invalid password", []Error{})
	ErrInvalidEmail    = NewAppError(401, "invalid email", []Error{})
	ErrInvalidUser     = NewAppError(401, "invalid user", []Error{})
	ErrInvalidRole     = NewAppError(401, "invalid role", []Error{})
	ErrInvalidScope    = NewAppError(401, "invalid scope", []Error{})
	ErrUserExists      = NewAppError(409, "user already exists", []Error{})

	// JWT errors
	ErrInvalidTokenFormat = NewAppError(401, "invalid token format", []Error{})
	ErrExpiredToken       = NewAppError(401, "token expired", []Error{})
	ErrInvalidTokenClaims = NewAppError(401, "invalid token claims", []Error{})
)
