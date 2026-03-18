// Package utils contains utility functions for the auth service.
package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
)

func StrIDToPgUUID(idStr string) (pgtype.UUID, *errors.AppError) {
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return pgtype.UUID{}, errors.NewAppError(errors.ErrBadRequest.Code, "bad request", []errors.Error{
			{Field: "id", Message: "The provided ID is not a valid UUID format."},
		})
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func FormatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Formats the 16-byte array into the standard UUID string format
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func GenerateRedisLoginKey(accountID, deviceID string) string {
	return fmt.Sprintf("%s:%s", accountID, deviceID)
}
