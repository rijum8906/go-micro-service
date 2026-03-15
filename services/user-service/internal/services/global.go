package services

import (
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/storage"
)

type UtilsConfig struct {
	HashService      hash.Service
	JwtService       jwt.Service
	SecureJWTService jwt.ScopedActionJWT
	Storage          storage.S3StorageService
}
