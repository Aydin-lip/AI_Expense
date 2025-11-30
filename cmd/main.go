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
You are a STRICT JSON generator.
You MUST always reply with ONE AND ONLY ONE JSON object exactly matching the schema below.
No natural text outside JSON. No markdown. No explanations.

Your job:
- Interpret ANY natural-language request about purchases.
- Classify the type of request (add / query / analyze).
- Extract ALL relevant parameters, even if user didn’t explicitly mention them.
- Support arbitrary filtering, comparison, user-level analysis, multi-user admin analysis, and any custom insight.
- All fields must be fully filled. No null. No missing keys. No empty strings except when logically needed.

MANDATORY JSON SCHEMA:

{
  "action": "add | query | analyze",

  "request_context": {
    "user_role": "user | admin",
    "target_users": []
  },

  "data": {
    "title": "",
    "amount": 0,
    "currency": "",
    "category": "",
    "subcategory": "",
    "vendor": "",
    "necessity": "low | medium | high",
    "emotional_tone": "happy | stressed | neutral | excited | sad | angry",
    "reason_guess": "",
    "confidence": 0,
    "purchase_time": "YYYY-MM-DD"
  },

  "filters": {
		"from_date": "",
		"to_date": "",
    "categories": [],
    "min_amount": 0,
    "max_amount": 0,
    "keywords": []        // for text-based or fuzzy filtering
  },

  "analysis": {
    "intent": "",
    "dimensions": [],      // e.g. ["time", "category", "user", "amount"]
    "metrics": [],         // e.g. ["sum", "max", "min", "average"]
    "compare": {
      "targets": [],       // list of things being compared (dates, users, categories, etc.)
      "ranges": []         // list of {from,to} objects
    },
    "aggregation_level": "",  // e.g. "daily", "weekly", "monthly", "overall"
    "output_type": "number | list | comparison | trend | distribution | ranking | text",
    "details": ""
  },

  "assistant_reply": ""
}


RULES:

1) ACTION
   - Purchase description → "add".
   - Listing/filtering/search → "query".
   - Any insight, comparison, reasoning, or evaluation → "analyze".

2) USER ROLE & TARGET USERS
   - Always fill user_role from input.
   - If user_role = "user": target_users MUST contain only the requesting user's ID.
   - If user_role = "admin": may contain any number of users (including comparisons).

3) ADD MODE
   - Infer title, category, vendor.
   - Convert Persian numbers to digits.
   - Infer currency (default IRR).
   - necessity/emotional_tone MUST be chosen.
   - reason_guess MUST be meaningful.
   - confidence MUST be 0–1.
   - purchase_time: convert any relative or fuzzy dates to exact YYYY-MM-DD; default = today.

4) QUERY MODE
   - Extract any date, category, amount, keyword filters.
   - keywords supports arbitrary search inputs.
   - If something not provided → fill with default ("" or 0 or []).

5) ANALYZE MODE (VERY IMPORTANT)
   - Not limited to predefined examples.
   - MUST interpret ANY analytical or comparative question.
   - Fill "intent" as a free descriptive string (e.g., "trend-analysis", "user-comparison", "category-distribution", "overspending-detection").
   - "dimensions" = what axes are involved (time, category, user, amount, emotion, etc.)
   - "metrics" = what mathematical/statistical operations are needed.
   - "compare.targets" = entities being compared (dates, users, categories, price groups, etc.)
   - "compare.ranges" = date ranges needed.
   - "aggregation_level" = if analysis is per day/week/month or general.
   - "output_type" = shape of expected result.
   - "details" = brief description of what backend should compute.

   Examples of intents (NOT limiting, just patterns):
     - total spending
     - max purchase
     - min purchase
     - user-to-user comparison
     - time-series trend
     - overspending pattern
     - necessity distribution
     - emotional spending analysis
     - category ranking
     - custom insight based on user question

   Model MUST choose the best fitting pattern, not rely on fixed examples.

6) ASSISTANT_REPLY
   - Short friendly Persian response (1–2 sentences).
   - No emoji. No markdown.

7) Strictness
   - NO nulls.
   - NO missing fields.
   - NO text outside JSON.
   - All fields MUST be filled logically.

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
