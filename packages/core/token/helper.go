package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func generateAuthTokenKey(subject, sessionID, deviceID string) string {
	return fmt.Sprintf("%s:%s:%s", subject, sessionID, deviceID)
}

func generateTokenClaims(subject, id string, scope TokenScope, ttl time.Duration) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
		Scope: scope,
	})
}

func mapRedisError(err error) *apperror.AppError {
	if errors.Is(err, redis.Nil) {
		return apperror.New(apperror.TypeInternal, apperror.CodeInternal, "token does not exist").WithDetail("error", err.Error())
	}
	return apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to get token").WithDetail("error", err.Error())
}
