package utils

import (
	"time"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StrToTokenScope(scope string) (corev1.TokenScope, bool) {
	s, ok := corev1.TokenScope_value[scope]
	return corev1.TokenScope(s), ok
}

func TimeToTimestamppb(t time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: int64(t.Second()),
	}
}
