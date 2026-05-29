package session

import (
	"context"

	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

// TODO: implement
func (s *SessionService) TerminateExpiredSessions(context.Context, *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	s.Logger.Error("Implement TerminateExpiredSessions")
	return nil, nil
}
