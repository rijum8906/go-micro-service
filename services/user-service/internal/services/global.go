package services

import (
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
)

type UtilsConfig struct {
	HashService hash.Service
	JwtService  jwt.Service
}
