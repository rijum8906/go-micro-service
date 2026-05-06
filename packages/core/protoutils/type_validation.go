package protoutils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

const (
	MaxPaginationLimit     = 100
	DefaultPaginationLimit = 10
)

// ValidatePaginationReq ensures the pagination parameters are within acceptable bounds.
func ValidatePaginationReq(req *corev1.PaginationRequest) *apperror.AppError {
	if req == nil {
		return apperror.ErrValidation.WithMessage("pagination request cannot be nil")
	}

	// Validate Page
	if req.Page < 1 {
		return apperror.ErrValidation.WithMessage("page number must be greater than 0")
	}

	// Validate and Sanitize Limit
	if req.Limit < 1 {
		return apperror.ErrValidation.WithMessage("limit must be at least 1")
	}

	if req.Limit > MaxPaginationLimit {
		return apperror.ErrValidation.WithMessage("limit cannot exceed 100")
	}

	return nil
}

// ValidateIDAndScopedTokenReq ensures the ID is a valid UUID and the scope is provided.
func ValidateIDAndScopedTokenReq(req *corev1.IDAndScopedTokenRequest) *apperror.AppError {
	if req == nil {
		return apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	// Validate UUID format
	if _, err := uuid.Parse(req.Id); err != nil {
		return apperror.ErrValidation.WithMessage("provided id is not a valid uuid")
	}

	// Validate Token Scope
	if !token.ValidateTokenScope(req.TokenScope) {
		return apperror.ErrValidation.WithMessage("token scope must be provided")
	}

	return nil
}
