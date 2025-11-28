package middleware

import (
	"strings"

	"example/AI/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Expected: "Bearer token..."
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid Authorization header"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		claims, err := utils.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// inject claims into context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)

		c.Next()
	}
}
