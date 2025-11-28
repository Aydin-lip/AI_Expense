package store

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var DB *gorm.DB

// dsn example: sqlserver://username:password@host:port?database=dbname
func Connect(dsn, dbName string) error {
	// Connect to the SQL Server
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(dsn)
		fmt.Println(err)
		panic("Failed to connect database (before create database)")
	}

	// Check if the database exists, and create it if it doesn't
	createDatabaseCommand := fmt.Sprintf("IF DB_ID('%s') IS NULL CREATE DATABASE [%s]", dbName, dbName)
	if err := db.Exec(createDatabaseCommand).Error; err != nil {
		fmt.Println("Error creating database:", err)
	}

	// Connect to the SQL Server database
	dsn = fmt.Sprintf("%s?database=%s", dsn, dbName)
	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		panic("Failed to connect database (after create database)")
	}

	// automigrate
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	DB = db

	// Seed admin user if not exists
	var admin User
	if err := DB.Where("username = ?", "admin").First(&admin).Error; err != nil {
		// not found -> create
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		admin = User{
			Username:     "admin",
			PasswordHash: string(hashed),
			Role:         "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Println("Admin user created: username='admin', password='admin'")
	} else {
		fmt.Println("Admin user already exists")
	}

	return nil
}
