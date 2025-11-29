package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"example/AI/internal/handlers"
	"example/AI/internal/middleware"
	"example/AI/internal/services"
	"example/AI/internal/store"
)

const dbName = "AI_Expense"
const dsn_static = "sqlserver://ArvinRayanSystem:Aria2661326441@localhost:1433"

// const secret_static = "Aydin_TesT"

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func main() {
	// load env vars (simple)
	dsn, exists := os.LookupEnv("DB_DSN")
	if !exists {
		// log.Fatal("DB_DSN env var is required (e.g. sqlserver://user:pass@host:port?database=dbname)")
		dsn = dsn_static
	}

	if err := store.Connect(dsn, dbName); err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	r := gin.Default()

	// auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterHandler)
		auth.POST("/login", handlers.LoginHandler)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthRequired()) // ⬅️ همین خط مهمه
	{
		api.GET("/me", handlers.MeHandler)
	}

	// Your ONLY reply must be inside:
	// <system_output>
	// { ...json... }
	// </system_output>
	prompt := `
You act as a command classifier AND a structured data extractor AND an AI assistant.

You MUST return JSON with ALL FIELDS FILLED — no empty, no missing fields, no null.

JSON STRUCTURE (MANDATORY):

{
  "action": "add | query | analyze",
  "data": {
    "title": "",
    "amount": 0,
    "currency": "",
    "category": "",
    "subcategory": "",
    "vendor": "",
    "necessity": "",
    "emotional_tone": "",
    "reason_guess": "",
    "confidence": 0,
    "purchase_time": "YYYY-MM-DD"
  },
  "filters": {},
  "assistant_reply": ""
}

RULES:
- If user describes a purchase → action="add" AND fill all "data" fields.
- Infer missing info logically (even if user didn't say it).
- Dates: convert “today / yesterday / last week” into exact date.
- Money: convert words into exact number.
- emotional_tone: infer emotion (happy, stressed, neutral, etc.)
- necessity: must be "low", "medium", or "high".
- confidence: number 0–1 estimating certainty.
- assistant_reply: a short friendly natural language response the assistant would say after saving data.
- No natural text outside assistant_reply.
- Never omit or leave any field empty.
`

	aiService := services.NewAIServiceFromEnv(prompt) // systemPrompt = همان سیستم پرامپت
	purchaseRepo := store.NewPurchaseRepo(store.DB)
	purchaseSvc := services.NewPurchaseService(purchaseRepo)

	aiHandler := handlers.NewAiHandler(aiService, purchaseSvc, store.DB) // یا مستقیم db

	r.POST("/ai/message", middleware.AuthRequired(), aiHandler.HandleMessage())

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
