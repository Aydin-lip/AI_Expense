package handlers

import (
	"fmt"
	"net/http"
	"time"

	"example/AI/internal/models"
	"example/AI/internal/services"
	"example/AI/internal/utils"

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
		username, exists := c.Get("username")
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
			return
		}
		userID := uidVal.(int)

		// send to AI
		parsed, assistantText, err := h.AI.ProcessMessage(body.Message, userID, username.(string), role.(string))
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
			fmt.Println(parsed)
			pf := utils.ConvertAIFiltersToPurchaseFilter(parsed.Filters, parsed.RequestContext)
			fmt.Println("=================")
			fmt.Println("filter : ", pf)
			fmt.Println("=================")
			items, err := h.Purchase.Query(pf)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"message":   parsed.AssistantReply,
				"purchases": items,
				"total":     len(items),
			})
			return

		case "analyze":
			// parsed.Filters -> convert to PurchaseFilter
			pf := utils.ConvertAIFiltersToPurchaseFilter(parsed.Filters, parsed.RequestContext)

			// compute server-side analytics
			total, _ := h.Purchase.SumAmount(pf)
			cat, catTotal, _ := h.Purchase.TopCategory(pf)
			count, _ := h.Purchase.CountPurchases(pf) // اگر تابع Count نداری، بساز: SELECT COUNT(*)

			analysisPayload := map[string]interface{}{
				"total_amount":       total,
				"top_category":       cat,
				"top_category_total": catTotal,
				"purchase_count":     count,
				"filters":            parsed.Filters,
			}

			// generate friendly natural summary via AI
			natural, raw, err := h.AI.GenerateNaturalAnalysis(analysisPayload)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
					"raw":   raw,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  natural,         // natural Persian summary
				"analysis": analysisPayload, // raw numbers
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
