// Package coreutils
package coreutils

import (
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ParseToUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}

	return u, nil
}

func ParseToTimestamppb(t time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: int64(t.Second()),
	}
}
