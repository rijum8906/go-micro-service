package utils

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func ParseAuthMethod(method token.AuthMethod) (corev1.AuthMethod, *apperror.AppError) {
	m, ok := corev1.AuthMethod_value[string(method)]
	if !ok {
		return corev1.AuthMethod_AUTH_METHOD_UNSPECIFIED, apperror.ErrValidation
	}
	return corev1.AuthMethod(m), nil
}

func ParseScope(scope token.TokenScope) (corev1.TokenScope, *apperror.AppError) {
	s, ok := corev1.TokenScope_value[string(scope)]
	if !ok {
		return corev1.TokenScope_TOKEN_SCOPE_UNSPECIFIED, apperror.ErrValidation
	}
	return corev1.TokenScope(s), nil
}
