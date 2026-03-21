package utils

import (
	"context"

	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/packages/common/token"
)

func ValidateBearerToken(ctx context.Context, token string, jwtService token.TokenService) (*token.Claims, *errors.AppError) {
	claims, appErr := jwtService.ValidateTokenWithClaims(ctx, token)
	if appErr != nil {
		return nil, appErr
	}

	return claims, nil
}
