// Package app
package app

import "github.com/gin-gonic/gin"

func (s *server) CreateGroup(relativePath string) gin.IRoutes {
	routerGroup := s.router.Group(relativePath)
	return routerGroup
}
