package utils

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func QueryOne[T any](value T, err error, notFoundMessage, internalMessage string) (*T, *apperror.AppError) {
	if err == nil {
		return &value, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		if notFoundMessage == "" {
			notFoundMessage = "resource not found"
		}

		return nil, apperror.ErrNotFound.WithMessage(notFoundMessage)
	}

	if internalMessage == "" {
		internalMessage = "database query failed"
	}

	return nil, apperror.ErrInternal.WithMessage(internalMessage).WithDetail("error", err.Error())
}

func QueryMany[T any](values []T, err error, internalMessage string) ([]T, *apperror.AppError) {
	if err == nil {
		return values, nil
	}

	if internalMessage == "" {
		internalMessage = "database query failed"
	}

	return nil, apperror.ErrInternal.WithMessage(internalMessage).WithDetail("error", err.Error())
}
