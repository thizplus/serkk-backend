package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("📋 Adding display_name and avatar columns to user_profiles...")

	// Database connection
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Add display_name column if not exists
	err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_profiles' AND column_name = 'display_name'
			) THEN
				ALTER TABLE user_profiles ADD COLUMN display_name VARCHAR(100) NOT NULL DEFAULT '';
			END IF;
		END $$;
	`).Error

	if err != nil {
		log.Fatalf("Failed to add display_name column: %v", err)
	}

	log.Println("✅ display_name column added")

	// Add avatar column if not exists
	err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_profiles' AND column_name = 'avatar'
			) THEN
				ALTER TABLE user_profiles ADD COLUMN avatar VARCHAR(500) DEFAULT '';
			END IF;
		END $$;
	`).Error

	if err != nil {
		log.Fatalf("Failed to add avatar column: %v", err)
	}

	log.Println("✅ avatar column added")

	// Verify columns
	var result struct {
		DisplayNameExists bool
		AvatarExists      bool
	}

	db.Raw(`
		SELECT
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'display_name') as display_name_exists,
			EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'avatar') as avatar_exists
	`).Scan(&result)

	log.Printf("📊 Verification: display_name exists: %v, avatar exists: %v", result.DisplayNameExists, result.AvatarExists)

	log.Println("🎉 Done!")
}
