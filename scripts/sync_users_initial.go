package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User represents the user model in Auth Service
type AuthUser struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	Email       string    `gorm:"column:email"`
	Username    string    `gorm:"column:username"`
	DisplayName string    `gorm:"column:display_name"`
	Avatar      string    `gorm:"column:avatar"`
	Role        string    `gorm:"column:role"`
	IsActive    bool      `gorm:"column:is_active"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (AuthUser) TableName() string {
	return "users"
}

// UsersCache represents the cached user in Backend Service
type UsersCache struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	Email       string    `gorm:"column:email"`
	Username    string    `gorm:"column:username"`
	DisplayName string    `gorm:"column:display_name"`
	Avatar      string    `gorm:"column:avatar"`
	Role        string    `gorm:"column:role"`
	IsActive    bool      `gorm:"column:is_active"`
	SyncedAt    time.Time `gorm:"column:synced_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (UsersCache) TableName() string {
	return "users_cache"
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Database configurations
	authDBHost := getEnv("AUTH_DB_HOST", "localhost")
	authDBPort := getEnv("AUTH_DB_PORT", "5432")
	authDBName := getEnv("AUTH_DB_NAME", "gofiber_auth")
	authDBUser := getEnv("AUTH_DB_USER", "postgres")
	authDBPass := getEnv("AUTH_DB_PASSWORD", "")

	backendDBHost := getEnv("DB_HOST", "localhost")
	backendDBPort := getEnv("DB_PORT", "5432")
	backendDBName := getEnv("DB_NAME", "gofiber_social")
	backendDBUser := getEnv("DB_USER", "postgres")
	backendDBPass := getEnv("DB_PASSWORD", "")

	log.Println("🔄 Starting initial user sync...")
	log.Printf("📍 Auth DB: %s:%s/%s", authDBHost, authDBPort, authDBName)
	log.Printf("📍 Backend DB: %s:%s/%s", backendDBHost, backendDBPort, backendDBName)

	// Connect to Auth DB
	authDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		authDBHost, authDBUser, authDBPass, authDBName, authDBPort,
	)

	authDB, err := gorm.Open(postgres.Open(authDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Auth DB: %v", err)
	}
	log.Println("✅ Connected to Auth DB")

	// Connect to Backend DB
	backendDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		backendDBHost, backendDBUser, backendDBPass, backendDBName, backendDBPort,
	)

	backendDB, err := gorm.Open(postgres.Open(backendDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Backend DB: %v", err)
	}
	log.Println("✅ Connected to Backend DB")

	// Fetch all users from Auth DB
	var authUsers []AuthUser
	if err := authDB.Find(&authUsers).Error; err != nil {
		log.Fatalf("❌ Failed to fetch users from Auth DB: %v", err)
	}
	log.Printf("📊 Found %d users in Auth DB", len(authUsers))

	// Check existing users in cache
	var existingCount int64
	backendDB.Model(&UsersCache{}).Count(&existingCount)
	log.Printf("📊 Existing users in cache: %d", existingCount)

	// Sync users to Backend DB
	syncedCount := 0
	failedCount := 0
	now := time.Now()

	for i, authUser := range authUsers {
		cacheUser := UsersCache{
			ID:          authUser.ID,
			Email:       authUser.Email,
			Username:    authUser.Username,
			DisplayName: authUser.DisplayName,
			Avatar:      authUser.Avatar,
			Role:        authUser.Role,
			IsActive:    authUser.IsActive,
			SyncedAt:    now,
			CreatedAt:   authUser.CreatedAt,
			UpdatedAt:   authUser.UpdatedAt,
		}

		// Upsert (insert or update)
		err := backendDB.Exec(`
			INSERT INTO users_cache (id, email, username, display_name, avatar, role, is_active, synced_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				username = EXCLUDED.username,
				display_name = EXCLUDED.display_name,
				avatar = EXCLUDED.avatar,
				role = EXCLUDED.role,
				is_active = EXCLUDED.is_active,
				synced_at = EXCLUDED.synced_at,
				updated_at = EXCLUDED.updated_at
		`, cacheUser.ID, cacheUser.Email, cacheUser.Username, cacheUser.DisplayName,
			cacheUser.Avatar, cacheUser.Role, cacheUser.IsActive, cacheUser.SyncedAt,
			cacheUser.CreatedAt, cacheUser.UpdatedAt).Error

		if err != nil {
			log.Printf("❌ Failed to sync user %s (%s): %v", authUser.Username, authUser.ID, err)
			failedCount++
		} else {
			syncedCount++
			if (i+1)%100 == 0 {
				log.Printf("⏳ Synced %d/%d users...", i+1, len(authUsers))
			}
		}
	}

	log.Println("\n" + repeatString("=", 60))
	log.Println("✅ Sync completed!")
	log.Printf("📊 Total users: %d", len(authUsers))
	log.Printf("✅ Synced: %d", syncedCount)
	log.Printf("❌ Failed: %d", failedCount)
	log.Println(repeatString("=", 60))

	// Verify sync
	var finalCount int64
	backendDB.Model(&UsersCache{}).Count(&finalCount)
	log.Printf("\n📊 Final count in users_cache: %d", finalCount)

	if finalCount == int64(len(authUsers)) {
		log.Println("🎉 All users synced successfully!")
	} else {
		log.Printf("⚠️  Warning: Expected %d users, but have %d in cache", len(authUsers), finalCount)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper to repeat strings
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
