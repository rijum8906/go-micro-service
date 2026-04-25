// Package protoutils provides utilities for working with protobuf messages
package protoutils

import (
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapTimestamp(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}

	return timestamppb.New(ts.Time)
}
