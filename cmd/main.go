package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"example/AI/internal/handlers"
	"example/AI/internal/middleware"
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

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
