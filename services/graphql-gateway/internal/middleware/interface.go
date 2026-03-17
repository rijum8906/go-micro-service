// Package middleware contains the middleware interface
package middleware

import (
	"net/http"

	"github.com/rijum8906/relay/packages/common/env"
)

type Middleware interface {
	AuthMiddleware(next http.Handler) http.Handler
}

type middleware struct {
	env *env.Env
}

func NewService(env *env.Env) Middleware {
	return &middleware{
		env: env,
	}
}
