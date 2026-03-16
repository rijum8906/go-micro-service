// Package server
package server

import "github.com/gin-gonic/gin"

func (s *server) CreateGroup(relativePath string) *gin.RouterGroup {
	routerGroup := s.router.Group(relativePath)
	return routerGroup
}

func (s *server) AddHandler(routerGroup *gin.RouterGroup, options HandlerOptions) {
}
