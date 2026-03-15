package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			return
		}

		// 2. Check for "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization format",
			})
			return
		}

		tokenString := parts[1]

		// 3. Verify the token
		claims, err := m.services.JwtService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			return
		}

		// 4. Inject UserID into context using the typed key for safety
		c.Set(string(UserIDKey), claims.UserID)
		c.Next()
	}
}
