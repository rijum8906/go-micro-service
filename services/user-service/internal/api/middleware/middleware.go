// Package middleware contains middleware for the auth service.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/packages/common/hash"
	"github.com/rijum8906/relay/packages/common/jwt"
)

type Middleware interface {
	AuthMiddleware() gin.HandlerFunc
	RequestID() gin.HandlerFunc
	Logger() gin.HandlerFunc
	Recovery() gin.HandlerFunc
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

func (m *middleware) RequestID() gin.HandlerFunc {
	return RequestIDMiddleware()
}

func (m *middleware) Logger() gin.HandlerFunc {
	return LoggerMiddleware()
}

func (m *middleware) Recovery() gin.HandlerFunc {
	return RecoveryMiddleware()
}
