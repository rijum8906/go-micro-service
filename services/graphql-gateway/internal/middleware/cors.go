// Package middleware
package middleware

import (
	"net/http"

	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rs/cors"
)

func CORS(httpHandler http.Handler, config *env.Config) http.Handler {
	debugEnabled := config.AppEnv == "development"

	c := cors.New(cors.Options{
		AllowedOrigins:   config.CorsAllowedOrigins,
		AllowedHeaders:   config.CorsAllowedHeaders,
		AllowedMethods:   config.CorsAllowedMethods,
		AllowCredentials: true,
		Debug:            debugEnabled,
	})

	return c.Handler(httpHandler)
}
