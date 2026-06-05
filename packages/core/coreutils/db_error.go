package coreutils

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func ParseDBError(err error, stuff string) *apperror.AppError {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound.WithMessage(stuff + " not found")
	}
	return apperror.ErrInternal.WithMessage("failed to fetch "+stuff+" from posrgres").WithDetail("db_error", err.Error())
}
