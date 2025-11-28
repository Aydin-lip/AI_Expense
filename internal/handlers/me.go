package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MeHandler(c *gin.Context) {
	userID := c.GetInt("userID")
	role := c.GetString("role")
	username := c.GetString("username")

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"role":     role,
		"username": username,
	})
}
