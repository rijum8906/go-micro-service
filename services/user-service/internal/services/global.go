package services

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
)

type UtilsConfig struct {
	HashService hash.Service
	JwtService  jwt.Service
}

func FormatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Formats the 16-byte array into the standard UUID string format
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
