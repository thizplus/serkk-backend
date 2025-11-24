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

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	log.Println("🔧 Starting duplicate conversations cleanup...")

	// Load environment variables
	godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "gofiber_social")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Step 1: Find and delete duplicate conversations
	log.Println("\n📊 Finding duplicate conversations...")

	type DuplicateConversation struct {
		User1ID string
		User2ID string
		Count   int64
	}

	var duplicates []DuplicateConversation
	err = db.Raw(`
		SELECT user1_id, user2_id, COUNT(*) as count
		FROM conversations
		GROUP BY user1_id, user2_id
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error

	if err != nil {
		log.Fatalf("❌ Failed to find duplicates: %v", err)
	}

	if len(duplicates) == 0 {
		log.Println("✅ No duplicate conversations found!")
	} else {
		log.Printf("⚠️  Found %d duplicate conversation pairs\n", len(duplicates))

		// Delete duplicates, keeping only the oldest one
		for _, dup := range duplicates {
			log.Printf("   Cleaning up duplicates for users %s <-> %s (found %d conversations)\n",
				dup.User1ID, dup.User2ID, dup.Count)

			// Delete all but the oldest conversation
			result := db.Exec(`
				DELETE FROM conversations
				WHERE user1_id = ? AND user2_id = ?
				AND id NOT IN (
					SELECT id FROM conversations
					WHERE user1_id = ? AND user2_id = ?
					ORDER BY created_at ASC
					LIMIT 1
				)
			`, dup.User1ID, dup.User2ID, dup.User1ID, dup.User2ID)

			if result.Error != nil {
				log.Printf("   ❌ Failed to delete duplicates: %v\n", result.Error)
			} else {
				log.Printf("   ✅ Deleted %d duplicate conversation(s)\n", result.RowsAffected)
			}
		}
	}

	// Step 2: Ensure UNIQUE constraint exists
	log.Println("\n🔒 Checking UNIQUE constraint...")

	var constraintExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints
			WHERE table_name = 'conversations'
			AND constraint_name = 'conversations_user1_user2_unique'
			AND constraint_type = 'UNIQUE'
		)
	`).Scan(&constraintExists).Error

	if err != nil {
		log.Fatalf("❌ Failed to check constraint: %v", err)
	}

	if constraintExists {
		log.Println("✅ UNIQUE constraint already exists!")
	} else {
		log.Println("⚠️  UNIQUE constraint not found, creating it...")

		err = db.Exec(`
			ALTER TABLE conversations
			ADD CONSTRAINT conversations_user1_user2_unique
			UNIQUE (user1_id, user2_id)
		`).Error

		if err != nil {
			log.Fatalf("❌ Failed to create UNIQUE constraint: %v", err)
		}

		log.Println("✅ UNIQUE constraint created successfully!")
	}

	// Step 3: Verify final state
	log.Println("\n📊 Verifying final state...")

	var totalConversations int64
	var duplicateCount int64

	db.Raw("SELECT COUNT(*) FROM conversations").Scan(&totalConversations)
	db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT user1_id, user2_id, COUNT(*) as count
			FROM conversations
			GROUP BY user1_id, user2_id
			HAVING COUNT(*) > 1
		) as dups
	`).Scan(&duplicateCount)

	fmt.Printf("\n📈 Final Statistics:\n")
	fmt.Printf("   Total conversations: %d\n", totalConversations)
	fmt.Printf("   Duplicate pairs: %d\n", duplicateCount)

	if duplicateCount > 0 {
		log.Println("\n⚠️  Warning: Still found duplicates! Manual intervention may be required.")
	} else {
		log.Println("\n✅ All done! No duplicates remaining.")
	}
}
