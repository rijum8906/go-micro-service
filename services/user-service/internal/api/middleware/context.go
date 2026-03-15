package middleware

import "github.com/gin-gonic/gin"

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the key for user ID in context
	UserIDKey ContextKey = "user_id"
	// RequestIDKey is the key for request ID in context
	RequestIDKey ContextKey = "request_id"
)

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}

// GetRequestID retrieves the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	val, exists := c.Get(string(RequestIDKey))
	if !exists {
		return ""
	}
	requestID, _ := val.(string)
	return requestID
}
