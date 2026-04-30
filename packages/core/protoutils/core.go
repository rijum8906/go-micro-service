package protoutils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func ValidateIDAndScopedToken(req *corev1.IDAndScopedTokenRequest) error {
	if err := uuid.Validate(req.Id); err != nil {
		return apperror.ErrValidation.WithMessage("invalid id")
	}
	if !token.ValidateTokenScope(req.TokenScope) {
		return apperror.ErrValidation.WithMessage("invalid token scope")
	}

	return nil
}
