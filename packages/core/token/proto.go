package token

import (
	"strings"

	"github.com/rijum8906/relay/packages/core/apperror"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

// TokenScopeFromProto maps a protobuf TokenScope enum to the domain TokenScope string.
func TokenScopeFromProto(scope corev1.TokenScope) (TokenScope, *apperror.AppError) {
	name, ok := corev1.TokenScope_name[int32(scope)]
	if !ok || scope == corev1.TokenScope_TOKEN_SCOPE_UNSPECIFIED {
		return "", apperror.ErrValidation.WithMessage("invalid token scope")
	}
	ts := TokenScope(strings.TrimPrefix(name, "TOKEN_SCOPE_"))
	if !ts.Validate() {
		return "", apperror.ErrValidation.WithMessage("invalid token scope")
	}
	return ts, nil
}
