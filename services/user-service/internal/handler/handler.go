// Package handler
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

func RegisterHandlers(router *gin.RouterGroup, service services.AuthService) {
	router.POST("/signup", func(ctx *gin.Context) {
		context := ctx.Request.Context()
		var dto dto.SignUpDTO
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := service.SignUp(context, dto)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	})

	router.POST("/signin", func(ctx *gin.Context) {
		context := ctx.Request.Context()
		var dto dto.SignInDTO
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := service.Signin(context, dto)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	})
}
