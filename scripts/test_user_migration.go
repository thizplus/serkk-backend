package main

import (
	"fmt"
	"log"

	"gofiber-template/domain/models"
	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database")

	// Drop users table
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	fmt.Println("✓ Dropped users table")

	// Try to migrate ONLY User model
	fmt.Println("\nAttempting to migrate User model...")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("❌ Failed to migrate User model: %v", err)
	}

	fmt.Println("✓ User model migrated successfully!")

	// Try to select from users
	fmt.Println("\nAttempting SELECT * FROM users LIMIT 1...")
	var user models.User
	result := db.Limit(1).Find(&user)
	if result.Error != nil {
		log.Fatalf("❌ Failed to SELECT: %v", result.Error)
	}

	fmt.Println("✓ SELECT successful!")
	fmt.Println("\n✅ User model works correctly!")
}
