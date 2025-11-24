package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("📋 Dropping FK constraint from user_profiles...")

	// Database connection
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Drop FK constraint (since UserProfile model says "no DB FK due to separate databases")
	err = db.Exec(`
		ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS fk_user_profiles_user;
	`).Error

	if err != nil {
		log.Fatalf("Failed to drop FK constraint: %v", err)
	}

	log.Println("✅ FK constraint dropped (user_profiles is now free to reference any user table)")
	log.Println("🎉 Done!")
}
