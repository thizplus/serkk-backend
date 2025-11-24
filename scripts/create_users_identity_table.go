package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gofiber-template/domain/models"
)

func main() {
	log.Println("📋 Creating users_identity table...")

	// Database connection
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Auto migrate users_identity table
	if err := db.AutoMigrate(&models.UsersIdentity{}); err != nil {
		log.Fatalf("Failed to create users_identity table: %v", err)
	}

	log.Println("✅ users_identity table created successfully!")

	// Verify table exists
	var count int64
	db.Table("users_identity").Count(&count)
	log.Printf("📊 Table verification: %d rows in users_identity", count)

	log.Println("🎉 Done!")
}
