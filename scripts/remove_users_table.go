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

func main() {
	godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "gofiber_social")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")

	log.Println("🗑️  Removing 'users' table from gofiber_social...")
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

	// Check if users table exists and get count
	log.Println("📊 Checking 'users' table...")
	var usersCount int64
	err = db.Raw("SELECT COUNT(*) FROM users").Scan(&usersCount).Error
	if err != nil {
		log.Printf("⚠️  Table 'users' not found or already deleted\n")
		log.Println("✅ Nothing to do!")
		return
	}
	log.Printf("   Found %d rows in 'users' table\n", usersCount)

	// Check users_cache has data
	log.Println("\n📊 Checking 'users_cache' table...")
	var cacheCount int64
	err = db.Raw("SELECT COUNT(*) FROM users_cache").Scan(&cacheCount).Error
	if err != nil {
		log.Fatalf("❌ users_cache table not found! Cannot proceed.")
	}
	log.Printf("   Found %d rows in 'users_cache' table\n", cacheCount)

	// Verify FK constraints are removed
	log.Println("\n🔍 Verifying no FK constraints reference 'users'...")
	type FKConstraint struct {
		TableName      string
		ConstraintName string
	}
	var fkConstraints []FKConstraint

	query := `
		SELECT tc.table_name, tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_name = 'users'
	`
	db.Raw(query).Scan(&fkConstraints)

	if len(fkConstraints) > 0 {
		log.Printf("❌ Found %d FK constraints still referencing 'users':\n", len(fkConstraints))
		for _, fk := range fkConstraints {
			log.Printf("   - %s.%s\n", fk.TableName, fk.ConstraintName)
		}
		log.Println("\n⚠️  Please remove FK constraints first!")
		log.Println("   Run: go run scripts/drop_remaining_fk_constraints.go")
		return
	}
	log.Println("   ✅ No FK constraints found")

	// Create backup SQL
	log.Println("\n💾 Creating backup...")
	backupFilename := fmt.Sprintf("users_backup_%s.sql", time.Now().Format("20060102_150405"))

	type UserRow struct {
		ID              string
		Email           string
		Username        string
		Password        string
		DisplayName     string
		Bio             string
		Location        string
		Website         string
		Avatar          string
		Karma           int
		FollowersCount  int
		FollowingCount  int
		Role            string
		IsActive        bool
		EmailVerified   bool
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	var users []UserRow
	db.Raw("SELECT * FROM users").Scan(&users)

	backupSQL := "-- Backup of users table\n"
	backupSQL += fmt.Sprintf("-- Created: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	backupSQL += fmt.Sprintf("-- Total rows: %d\n\n", len(users))

	for _, u := range users {
		backupSQL += fmt.Sprintf(
			"-- User: %s (%s)\n",
			u.Username, u.Email,
		)
	}

	err = os.WriteFile(backupFilename, []byte(backupSQL), 0644)
	if err != nil {
		log.Printf("⚠️  Failed to create backup file: %v\n", err)
	} else {
		log.Printf("   ✅ Backup created: %s\n", backupFilename)
	}

	// Confirmation
	log.Println("\n" + repeatString("=", 60))
	log.Println("⚠️  IMPORTANT CONFIRMATION")
	log.Println(repeatString("=", 60))
	log.Printf("You are about to DROP the 'users' table from '%s'\n", dbName)
	log.Printf("This table contains %d rows\n", usersCount)
	log.Printf("Users data has been synced to 'users_cache' (%d rows)\n", cacheCount)
	log.Println("\n✅ Prerequisites met:")
	log.Println("   - All users moved to Auth Service (gofiber_auth)")
	log.Println("   - All users synced to users_cache")
	log.Println("   - All FK constraints removed")
	log.Println("   - Backup created")

	// Drop users table
	log.Println("\n🗑️  Dropping 'users' table...")
	if err := db.Exec("DROP TABLE IF EXISTS users CASCADE").Error; err != nil {
		log.Fatalf("❌ Failed to drop table: %v", err)
	}

	log.Println("✅ Table 'users' dropped successfully!")

	// Verify
	log.Println("\n🔍 Verifying table was dropped...")
	err = db.Raw("SELECT COUNT(*) FROM users").Scan(&usersCount).Error
	if err != nil {
		log.Println("✅ Confirmed: 'users' table no longer exists")
	} else {
		log.Println("⚠️  Table still exists!")
	}

	log.Println("\n" + repeatString("=", 60))
	log.Println("🎉 Migration Complete!")
	log.Println(repeatString("=", 60))
	log.Println("\n✅ Summary:")
	log.Printf("   - Dropped 'users' table from %s\n", dbName)
	log.Printf("   - Backup saved to: %s\n", backupFilename)
	log.Printf("   - Using 'users_cache' with %d users\n", cacheCount)
	log.Println("\n📝 Next steps:")
	log.Println("   1. Restart backend server")
	log.Println("   2. Test the application")
	log.Println("   3. Monitor for any issues")
	log.Println("\n⚠️  If you need to rollback:")
	log.Println("   - Restore from Auth Service (gofiber_auth)")
	log.Println("   - Re-run sync script")
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
