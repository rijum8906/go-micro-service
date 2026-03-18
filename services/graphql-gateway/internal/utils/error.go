package utils

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeBadUserInput        = "BAD_USER_INPUT"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeRateLimited         = "RATE_LIMITED"
	CodeTimeout             = "TIMEOUT"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeRequestCanceled     = "REQUEST_CANCELED"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
)

type AppError struct {
	Message string
	Code    string
	Err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewAppError(message, code string) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
	}
}

func WrapError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	grpcStatus, ok := status.FromError(err)
	if ok {
		return newAppErrorFromStatus(grpcStatus, err)
	}

	parsed := parseWrappedGRPCError(err.Error())
	if parsed != nil {
		parsed.Err = err
		return parsed
	}

	return NewAppError("Internal server error", CodeInternalServerError)
}

func newAppErrorFromStatus(grpcStatus *status.Status, err error) *AppError {
	return &AppError{
		Message: sanitizeClientMessage(grpcStatus.Message()),
		Code:    mapGRPCCode(grpcStatus.Code()),
		Err:     err,
	}
}

func PresentError(ctx context.Context, err error, exposeInternal bool) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	appErr := resolveAppError(err, gqlErr)

	if appErr == nil {
		gqlErr.Extensions = mergeExtensions(gqlErr.Extensions, map[string]any{
			"code": CodeInternalServerError,
		})
		return gqlErr
	}

	if appErr.Code == "" {
		appErr.Code = CodeInternalServerError
	}

	if appErr.Message != "" && (appErr.Code != CodeInternalServerError || exposeInternal) {
		gqlErr.Message = appErr.Message
	} else if appErr.Code == CodeInternalServerError && !exposeInternal {
		gqlErr.Message = "Internal server error"
	}

	gqlErr.Extensions = mergeExtensions(gqlErr.Extensions, map[string]any{
		"code": appErr.Code,
	})

	return gqlErr
}

func resolveAppError(err error, gqlErr *gqlerror.Error) *AppError {
	appErr := WrapError(err)

	if appErr != nil && appErr.Code != CodeInternalServerError {
		return appErr
	}

	if gqlErr == nil {
		return appErr
	}

	parsed := parseWrappedGRPCError(gqlErr.Message)
	if parsed != nil {
		parsed.Err = err
		return parsed
	}

	if appErr != nil {
		appErr.Message = sanitizeClientMessage(gqlErr.Message)
		return appErr
	}

	return &AppError{
		Message: sanitizeClientMessage(gqlErr.Message),
		Code:    CodeInternalServerError,
		Err:     err,
	}
}

func mergeExtensions(base map[string]any, extra map[string]any) map[string]any {
	if len(base) == 0 {
		base = make(map[string]any, len(extra))
	}

	for key, value := range extra {
		base[key] = value
	}

	return base
}

func mapGRPCCode(code codes.Code) string {
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return CodeBadUserInput
	case codes.Unauthenticated:
		return CodeUnauthorized
	case codes.PermissionDenied:
		return CodeForbidden
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists, codes.Aborted:
		return CodeConflict
	case codes.ResourceExhausted:
		return CodeRateLimited
	case codes.DeadlineExceeded:
		return CodeTimeout
	case codes.Canceled:
		return CodeRequestCanceled
	case codes.Unimplemented:
		return CodeNotImplemented
	case codes.Unavailable:
		return CodeServiceUnavailable
	default:
		return CodeInternalServerError
	}
}
