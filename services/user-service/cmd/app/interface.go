package app

import (
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/packages/common/env"
)

type Server interface {
	CreateGroup(relativePath string) gin.IRoutes
}

type server struct {
	env    *env.Env
	router gin.IRouter
}

func NewServer(env *env.Env) Server {
	router := gin.New()

	return &server{
		env:    env,
		router: router,
	}
}
