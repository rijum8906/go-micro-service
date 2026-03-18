package handlers

import (
	commonerrors "github.com/rijum8906/relay/packages/common/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var appErrorCodeToGRPC = map[string]codes.Code{
	commonerrors.ErrBadRequest.Code:         codes.InvalidArgument,
	commonerrors.ErrInvalidBody.Code:        codes.InvalidArgument,
	commonerrors.ErrInvalidInput.Code:       codes.InvalidArgument,
	commonerrors.ErrMissingField.Code:       codes.InvalidArgument,
	commonerrors.ErrValidation.Code:         codes.InvalidArgument,
	commonerrors.ErrUnauthorized.Code:       codes.Unauthenticated,
	commonerrors.ErrInvalidCredentials.Code: codes.Unauthenticated,
	commonerrors.ErrInvalidToken.Code:       codes.Unauthenticated,
	commonerrors.ErrInvalidTokenFormat.Code: codes.Unauthenticated,
	commonerrors.ErrExpiredToken.Code:       codes.Unauthenticated,
	commonerrors.ErrInvalidTokenClaims.Code: codes.Unauthenticated,
	commonerrors.ErrInvalidPassword.Code:    codes.Unauthenticated,
	commonerrors.ErrInvalidEmail.Code:       codes.InvalidArgument,
	commonerrors.ErrInvalidUser.Code:        codes.NotFound,
	commonerrors.ErrInvalidScope.Code:       codes.PermissionDenied,
	commonerrors.ErrForbidden.Code:          codes.PermissionDenied,
	commonerrors.ErrInvalidRole.Code:        codes.PermissionDenied,
	commonerrors.ErrNotFound.Code:           codes.NotFound,
	commonerrors.ErrDBNotFound.Code:         codes.NotFound,
	commonerrors.ErrConflict.Code:           codes.AlreadyExists,
	commonerrors.ErrUserExists.Code:         codes.AlreadyExists,
	commonerrors.ErrDBConflict.Code:         codes.AlreadyExists,
	commonerrors.ErrDBConnection.Code:       codes.Unavailable,
	commonerrors.ErrDBError.Code:            codes.Internal,
	commonerrors.ErrInternal.Code:           codes.Internal,
	"SESSION_NOT_FOUND":                     codes.NotFound,
}

func appErrorToGRPC(appErr *commonerrors.AppError) error {
	if appErr == nil {
		return nil
	}

	grpcCode, ok := appErrorCodeToGRPC[appErr.Code]
	if !ok {
		grpcCode = codes.Internal
	}

	return status.Error(grpcCode, appErr.Message)
}
