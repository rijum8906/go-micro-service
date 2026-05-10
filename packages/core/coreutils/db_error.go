package coreutils

import (
	"database/sql"
	"errors"

	"github.com/rijum8906/relay/packages/core/apperror"
)

func ParseDBError(err error, stuff string) *apperror.AppError {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.ErrNotFound.WithMessage(stuff + " not found")
	}
	return apperror.ErrInternal.WithMessage("failed to fetch membership").WithDetail("db_error", err.Error())
}
