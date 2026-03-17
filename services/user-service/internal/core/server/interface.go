package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/packages/common/env"
)

const (
	MethodGET = "GET"
)

type Server interface {
	CreateGroup(relativePath string) *gin.RouterGroup
}

type server struct {
	env    *env.Env
	router gin.IRouter
}

type HandlerOptions struct {
	path string
}

func NewServer(env *env.Env) Server {
	router := gin.New()

	return &server{
		env:    env,
		router: router,
	}
}
