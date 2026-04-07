package utils

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapAppError(appErr *apperror.AppError) error {
	if appErr == nil {
		return nil
	}

	if appErr.Code == apperror.CodeInternal {
		fmt.Println("error: ", appErr.Error()) // TODO: log.Error()
		fmt.Println("error details: ", appErr.Details)
	}

	switch appErr.Code {
	case apperror.CodeValidation:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case apperror.CodeUnAuthenticated:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperror.CodeForbidden:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case apperror.CodeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case apperror.CodeThirdParty:
		return status.Error(codes.Unavailable, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
