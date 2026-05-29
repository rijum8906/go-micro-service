package constants

import "github.com/rijum8906/relay/packages/core/apperror"

var (
	// Create fresh AppError instances so details don't accumulate on the shared ErrInternal
	ErrUserNotFoundInCtx       = apperror.New(apperror.CodeInternal, "Internal Server Error").WithDetail("internal_message", "failed to retrieve user info from context")
	ErrClientNotFoundInCtx     = apperror.New(apperror.CodeInternal, "Internal Server Error").WithDetail("internal_message", "failed to retrieve client info from context")
	ErrInvalidUserIDInUserInfo = apperror.New(apperror.CodeInternal, "Internal Server Error").WithDetail("internal_message", "invalid user id in user info received from context")
)
