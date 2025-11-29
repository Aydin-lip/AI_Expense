package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"example/AI/internal/models"
	"example/AI/internal/store"
	"example/AI/internal/utils"
)

// DTOs
type registerReq struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type authResp struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Handlers
func RegisterHandler(c *gin.Context) {
	var body registerReq
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{"error": "invalid payload"})
		return
	}

	// check existing user
	var existing models.User
	if err := store.DB.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		c.JSON(409, gin.H{"error": "username already taken"})
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not hash password"})
		return
	}

	user := models.User{
		Username:     body.Username,
		PasswordHash: string(hashed),
		Role:         "user",
	}

	if err := store.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to create user"})
		return
	}

	// generate token (short-lived so user can proceed in demo)
	token, err := utils.GenerateToken(int(user.ID), user.Role, user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(201, gin.H{
		"message":    "user created",
		"token":      token,
		"expires_at": time.Now().Add(24 * 5 * time.Hour),
	})
}

func LoginHandler(c *gin.Context) {
	var body loginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid payload"})
		return
	}

	var user models.User
	if err := store.DB.Where("username = ?", body.Username).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(404, gin.H{"error": "invalid credentials"})
		return
	}

	fmt.Println(user)
	token, err := utils.GenerateToken(int(user.ID), user.Role, user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(200, gin.H{
		"message":    "logged in",
		"token":      token,
		"expires_at": time.Now().Add(24 * 5 * time.Hour),
	})
}
