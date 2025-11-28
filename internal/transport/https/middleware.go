package https

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *HTTPHandlers) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			return
		}
		isBlacklisted, err := h.AuthService.IsTokenBlacklisted(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check token status"})
			return
		}
		if isBlacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}
		claims, err := h.AuthService.ParseTokenAndGetClaims(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		ctx := context.WithValue(c.Request.Context(), UserCtxKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
