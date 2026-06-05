package utils

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func AssertRowExists(err error, stuff, subject string) *apperror.AppError {
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrInternal.
				WithMessage(fmt.Sprintf("%s not found. Please contact support", stuff)).
				WithDetail("internal_message", fmt.Sprintf("%s not found in database", stuff)).
				WithDetail("subject", subject)
		}

		return apperror.ErrNotFound.
			WithMessage(fmt.Sprintf("%s not found", stuff))
	}
	return nil
}
