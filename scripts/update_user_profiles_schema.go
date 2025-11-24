package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gofiber-template/domain/models"
)

func main() {
	log.Println("📋 Updating user_profiles table schema...")

	// Database connection
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Auto migrate user_profiles table (will add missing columns)
	if err := db.AutoMigrate(&models.UserProfile{}); err != nil {
		log.Fatalf("Failed to update user_profiles table: %v", err)
	}

	log.Println("✅ user_profiles table updated successfully!")

	// Verify table schema
	var count int64
	db.Table("user_profiles").Count(&count)
	log.Printf("📊 Table verification: %d rows in user_profiles", count)

	log.Println("🎉 Done!")
}
