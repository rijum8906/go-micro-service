package coreutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ParseJWTTimestamp(ttl time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(ttl))
}
