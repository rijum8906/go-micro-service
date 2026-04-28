package utils

import (
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ParseOptionalUUID(value, field string) (pgtype.UUID, *apperror.AppError) {
	if strings.TrimSpace(value) == "" {
		return pgtype.UUID{}, nil
	}

	id, appErr := NewUUID(value)
	if appErr != nil {
		return pgtype.UUID{}, appErr.WithDetail("field", field)
	}

	return PGUUID(id), nil
}

func ParseOptionalTimestamp(value *timestamppb.Timestamp, field string) (pgtype.Timestamptz, *apperror.AppError) {
	if value == nil {
		return pgtype.Timestamptz{}, nil
	}

	if err := value.CheckValid(); err != nil {
		return pgtype.Timestamptz{}, apperror.ErrValidation.WithMessage("invalid timestamp").WithDetail("field", field).WithDetail("error", err.Error())
	}

	return pgtype.Timestamptz{
		Time:  value.AsTime(),
		Valid: true,
	}, nil
}

func PGUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: value,
		Valid: true,
	}
}

func UUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}

	return uuid.UUID(value.Bytes).String()
}

func Timestamp(value pgtype.Timestamptz) *timestamppb.Timestamp {
	if !value.Valid {
		return nil
	}

	return timestamppb.New(value.Time)
}
