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
	godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "gofiber_social")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")

	log.Println("🔍 Checking Database Status...")
	log.Printf("📍 Database: %s:%s/%s\n", dbHost, dbPort, dbName)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	log.Println("✅ Connected\n")

	// Check if users table exists
	log.Println("📊 Checking 'users' table:")
	var usersCount int64
	err = db.Raw("SELECT COUNT(*) FROM users").Scan(&usersCount).Error
	if err != nil {
		log.Printf("   ⚠️  Table 'users' not found or error: %v\n", err)
	} else {
		log.Printf("   ✅ Table 'users' exists with %d rows\n", usersCount)
	}

	// Check if users_cache table exists
	log.Println("\n📊 Checking 'users_cache' table:")
	var usersCacheCount int64
	err = db.Raw("SELECT COUNT(*) FROM users_cache").Scan(&usersCacheCount).Error
	if err != nil {
		log.Printf("   ❌ Table 'users_cache' not found: %v\n", err)
	} else {
		log.Printf("   ✅ Table 'users_cache' exists with %d rows\n", usersCacheCount)
	}

	// Check if users table has any FK constraints
	log.Println("\n🔗 Checking Foreign Key Constraints on 'users':")
	type FKConstraint struct {
		TableName      string
		ConstraintName string
		ColumnName     string
	}
	var fkConstraints []FKConstraint

	query := `
		SELECT
			tc.table_name,
			tc.constraint_name,
			kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_name = 'users'
		ORDER BY tc.table_name
	`

	err = db.Raw(query).Scan(&fkConstraints).Error
	if err != nil {
		log.Printf("   ❌ Error checking FK constraints: %v\n", err)
	} else if len(fkConstraints) > 0 {
		log.Printf("   ⚠️  Found %d FK constraints still referencing 'users':\n", len(fkConstraints))
		for _, fk := range fkConstraints {
			log.Printf("      - %s.%s (%s)\n", fk.TableName, fk.ColumnName, fk.ConstraintName)
		}
	} else {
		log.Println("   ✅ No FK constraints referencing 'users' (all removed)")
	}

	// Summary
	log.Println("\n" + repeatString("=", 60))
	log.Println("📋 Summary:")
	log.Println(repeatString("=", 60))

	if usersCount > 0 {
		log.Printf("⚠️  Old 'users' table still exists (%d rows)\n", usersCount)
		log.Println("   → Safe to keep for now (as backup)")
		log.Println("   → Can be deleted later after confirming everything works")
	}

	if usersCacheCount > 0 {
		log.Printf("✅ New 'users_cache' table ready (%d rows)\n", usersCacheCount)
	}

	log.Println("\n💡 Recommendations:")
	log.Println("   1. Keep 'users' table as backup (don't delete yet)")
	log.Println("   2. Use 'users_cache' for all new queries")
	log.Println("   3. Test the application thoroughly")
	log.Println("   4. Delete 'users' table only after 1-2 weeks of stable operation")
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
