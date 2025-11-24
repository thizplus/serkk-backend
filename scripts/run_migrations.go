package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "gofiber_social")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")

	log.Println("🔄 Running migrations...")
	log.Printf("📍 Database: %s:%s/%s", dbHost, dbPort, dbName)

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Migration 023: Create users_cache table
	log.Println("\n📝 Running migration 023: Create users_cache table")

	sql023 := `
CREATE TABLE IF NOT EXISTS users_cache (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    synced_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_cache_email ON users_cache(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_cache_username ON users_cache(username);
CREATE INDEX IF NOT EXISTS idx_users_cache_synced_at ON users_cache(synced_at);
CREATE INDEX IF NOT EXISTS idx_users_cache_is_active ON users_cache(is_active);
`

	if err := db.Exec(sql023).Error; err != nil {
		log.Fatalf("❌ Failed to run migration 023: %v", err)
	}
	log.Println("✅ Migration 023 completed")

	// Migration 024: Drop users FK constraints
	log.Println("\n📝 Running migration 024: Drop users FK constraints")

	sql024 := `
-- Drop FK constraints from posts table
ALTER TABLE IF EXISTS posts
    DROP CONSTRAINT IF EXISTS fk_posts_author,
    DROP CONSTRAINT IF EXISTS posts_author_id_fkey;

-- Drop FK constraints from comments table
ALTER TABLE IF EXISTS comments
    DROP CONSTRAINT IF EXISTS fk_comments_author,
    DROP CONSTRAINT IF EXISTS comments_author_id_fkey;

-- Drop FK constraints from follows table
ALTER TABLE IF EXISTS follows
    DROP CONSTRAINT IF EXISTS fk_follows_follower,
    DROP CONSTRAINT IF EXISTS follows_follower_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_follows_following,
    DROP CONSTRAINT IF EXISTS follows_following_id_fkey;

-- Drop FK constraints from notifications table
ALTER TABLE IF EXISTS notifications
    DROP CONSTRAINT IF EXISTS fk_notifications_user,
    DROP CONSTRAINT IF EXISTS notifications_user_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_notifications_sender,
    DROP CONSTRAINT IF EXISTS notifications_sender_id_fkey;

-- Drop FK constraints from votes table
ALTER TABLE IF EXISTS votes
    DROP CONSTRAINT IF EXISTS fk_votes_user,
    DROP CONSTRAINT IF EXISTS votes_user_id_fkey;

-- Drop FK constraints from messages table
ALTER TABLE IF EXISTS messages
    DROP CONSTRAINT IF EXISTS fk_messages_sender,
    DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;

-- Drop FK constraints from saved_posts table
ALTER TABLE IF EXISTS saved_posts
    DROP CONSTRAINT IF EXISTS fk_saved_posts_user,
    DROP CONSTRAINT IF EXISTS saved_posts_user_id_fkey;

-- Drop FK constraints from blocks table
ALTER TABLE IF EXISTS blocks
    DROP CONSTRAINT IF EXISTS fk_blocks_blocker,
    DROP CONSTRAINT IF EXISTS blocks_blocker_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_blocks_blocked,
    DROP CONSTRAINT IF EXISTS blocks_blocked_id_fkey;
`

	if err := db.Exec(sql024).Error; err != nil {
		log.Fatalf("❌ Failed to run migration 024: %v", err)
	}
	log.Println("✅ Migration 024 completed")

	// Verify users_cache table
	log.Println("\n📊 Verifying users_cache table...")
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM users_cache").Scan(&count).Error; err != nil {
		log.Fatalf("❌ Failed to query users_cache: %v", err)
	}
	log.Printf("✅ users_cache table exists with %d rows", count)

	log.Println("\n============================================================")
	log.Println("🎉 All migrations completed successfully!")
	log.Println("============================================================")
	log.Println("\nNext steps:")
	log.Println("1. Run initial user sync: go run scripts/sync_users_initial.go")
	log.Println("2. Restart the backend server")
	log.Println("3. Test the application")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
