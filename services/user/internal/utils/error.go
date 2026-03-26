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

	if appErr.Type == apperror.TypeInternal {
		fmt.Println("error: ", appErr.Error()) // TODO: log.Error()
		fmt.Println("error details: ", appErr.Details)
	}

	switch appErr.Type {
	case apperror.TypeValidation:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case apperror.TypeUnAuthenticated:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperror.TypeForbidden:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case apperror.TypeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case apperror.TypeThirdParty:
		return status.Error(codes.Unavailable, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
