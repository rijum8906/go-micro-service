package coreutils

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ParseToProtoTimestamp(t time.Duration) *timestamppb.Timestamp {
	return timestamppb.New(time.Now().Add(t))
}
