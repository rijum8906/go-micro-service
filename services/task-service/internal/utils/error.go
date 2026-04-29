package utils

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapAppError(appErr *apperror.AppError) error {
	if appErr == nil {
		return nil
	}

	switch appErr.Code {
	case apperror.CodeValidation:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case apperror.CodeUnAuthenticated:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperror.CodeForbidden:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case apperror.CodeConflict:
		return status.Error(codes.AlreadyExists, appErr.Message)
	case apperror.CodeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case apperror.CodeThirdParty:
		return status.Error(codes.Unavailable, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
