package handlers

import (
	"net/http"
	"time"

	"example/AI/internal/models"
	"example/AI/internal/services"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

type AiMessageReq struct {
	Message string `json:"message" binding:"required"`
}

type AiHandler struct {
	AI       *services.AIService
	Purchase *services.PurchaseService
	DB       *gorm.DB // or *gorm.DB
}

func NewAiHandler(ai *services.AIService, ps *services.PurchaseService, dbw *gorm.DB) *AiHandler {
	return &AiHandler{AI: ai, Purchase: ps, DB: dbw}
}

func (h *AiHandler) HandleMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body AiMessageReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// get user from context (middleware set)
		uidVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
			return
		}
		userID := uidVal.(int)

		// send to AI
		parsed, assistantText, err := h.AI.ProcessMessage(body.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "raw_ai": assistantText})
			return
		}

		// save ai log
		aiLog := models.AILog{
			InputText: body.Message,
			AIOutput:  assistantText,
			CreatedAt: time.Now().UTC(),
		}
		// attach user id when available
		aiLog.UserID = &userID
		_ = h.DB.Create(&aiLog) // ignoring error for brevity

		// handle action
		switch parsed.Action {
		case "create_purchase", "add":
			aiData := parsed.Data["extracted"]
			// in our earlier schema we put extracted inside data.extracted
			var mapData map[string]interface{}
			if m, ok := aiData.(map[string]interface{}); ok {
				mapData = m
			} else {
				// maybe data itself is the object
				mapData = parsed.Data
			}

			p, err := h.Purchase.CreateFromAIData(userID, mapData)
			if err != nil {
				// reply natural-language assistantText + an error
				c.JSON(http.StatusBadRequest, gin.H{
					"message": assistantText,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  parsed.AssistantReply,
				"purchase": p,
			})
			return

		case "get_purchases", "query":
			// convert parsed.Filters => PurchaseFilter (we didn't define full mapping here)
			// For MVP: if filters contain from/to and maybe categories, map them
			// ... implement mapping util ConvertAIFiltersToPurchaseFilter(...)
			c.JSON(http.StatusOK, gin.H{
				"message": assistantText,
				"data":    parsed.Filters,
			})
			return

		case "analyze":
			c.JSON(http.StatusOK, gin.H{
				"message":  assistantText,
				"analysis": parsed.Data,
			})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"message": assistantText,
				"error":   "unknown action",
			})
			return
		}
	}
}
