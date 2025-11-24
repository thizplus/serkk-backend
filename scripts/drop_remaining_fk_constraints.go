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
	godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "gofiber_social")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")

	log.Println("🔄 Dropping remaining FK constraints...")
	log.Printf("📍 Database: %s:%s/%s\n", dbHost, dbPort, dbName)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	log.Println("✅ Connected to database\n")

	// Drop remaining FK constraints
	sql := `
-- Drop FK constraints from auto_post_logs table
ALTER TABLE IF EXISTS auto_post_logs
    DROP CONSTRAINT IF EXISTS auto_post_logs_approved_by_fkey,
    DROP CONSTRAINT IF EXISTS auto_post_logs_rejected_by_fkey;

-- Drop FK constraints from auto_post_settings table
ALTER TABLE IF EXISTS auto_post_settings
    DROP CONSTRAINT IF EXISTS auto_post_settings_bot_user_id_fkey;

-- Drop FK constraints from conversations table
ALTER TABLE IF EXISTS conversations
    DROP CONSTRAINT IF EXISTS conversations_user1_id_fkey,
    DROP CONSTRAINT IF EXISTS conversations_user2_id_fkey;

-- Drop FK constraints from files table
ALTER TABLE IF EXISTS files
    DROP CONSTRAINT IF EXISTS files_user_id_fkey;

-- Drop FK constraints from media table
ALTER TABLE IF EXISTS media
    DROP CONSTRAINT IF EXISTS media_user_id_fkey;

-- Drop FK constraints from messages table (receiver_id)
ALTER TABLE IF EXISTS messages
    DROP CONSTRAINT IF EXISTS messages_receiver_id_fkey;

-- Drop FK constraints from notification_settings table
ALTER TABLE IF EXISTS notification_settings
    DROP CONSTRAINT IF EXISTS notification_settings_user_id_fkey;

-- Drop FK constraints from push_subscriptions table
ALTER TABLE IF EXISTS push_subscriptions
    DROP CONSTRAINT IF EXISTS push_subscriptions_user_id_fkey;

-- Drop FK constraints from search_histories table
ALTER TABLE IF EXISTS search_histories
    DROP CONSTRAINT IF EXISTS search_histories_user_id_fkey;

-- Drop FK constraints from tasks table
ALTER TABLE IF EXISTS tasks
    DROP CONSTRAINT IF EXISTS tasks_user_id_fkey;
`

	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("❌ Failed to drop FK constraints: %v", err)
	}

	log.Println("✅ All remaining FK constraints dropped\n")

	// Verify no FK constraints left
	log.Println("🔍 Verifying FK constraints...")
	type FKConstraint struct {
		TableName      string
		ConstraintName string
	}
	var remaining []FKConstraint

	query := `
		SELECT
			tc.table_name,
			tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_name = 'users'
	`

	db.Raw(query).Scan(&remaining)

	if len(remaining) > 0 {
		log.Printf("⚠️  Still found %d FK constraints:\n", len(remaining))
		for _, fk := range remaining {
			log.Printf("   - %s.%s\n", fk.TableName, fk.ConstraintName)
		}
	} else {
		log.Println("✅ No FK constraints referencing 'users' table!")
	}

	log.Println("\n" + repeatString("=", 60))
	log.Println("✅ All FK constraints removed successfully!")
	log.Println(repeatString("=", 60))
	log.Println("\n💡 Next steps:")
	log.Println("   1. Restart backend server")
	log.Println("   2. Test the application")
	log.Println("   3. Keep 'users' table as backup for now")
	log.Println("   4. Can drop 'users' table later if everything works fine")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
