package user

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
)

func valiateGenerateScopeTokenReq(req *userv1.GenerateScopedTokenRequest) *apperror.AppError {
	if req == nil {
		return apperror.ErrValidation.WithMessage("generate scoped token request is required")
	}

	if req.AuthMethod.String() != string(coreconstants.AuthMethodPassword) {
		return apperror.ErrValidation.WithMessage("invalid auth method")
	}

	if !constants.IsValidaTokenScope(req.Scope.String()) {
		return apperror.ErrValidation.WithMessage("invalid token scope")
	}
	return nil
}

func validateChangePasswordReq(req *userv1.ChangePasswordRequest) *apperror.AppError {
	if req == nil {
		return apperror.ErrValidation.WithMessage("change password request is required")
	}

	if req.TokenScope != string(constants.TokenScopeChangePassword) {
		return apperror.ErrValidation.WithMessage("invalid token scope")
	}

	return nil
}
