// Package middleware contains middleware for the user service.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
)

type Middleware interface {
	AuthMiddleware() gin.HandlerFunc
}

type middleware struct {
	services Services
}

type Services struct {
	HashService hash.Service
	JwtService  jwt.Service
}

func NewMiddleware(utils Services) Middleware {
	return &middleware{
		services: utils,
	}
}
