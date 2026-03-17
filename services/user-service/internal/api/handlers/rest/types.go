package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
)

type Handler interface{}

type handler struct {
	middlewares []middleware.Middleware
	router      gin.IRoutes
}

func NewHandler(router gin.IRoutes, middlewares []middleware.Middleware) Handler {
	return &handler{
		router:      router,
		middlewares: middlewares,
	}
}
