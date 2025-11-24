package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("📊 Creating user_profiles table...")

	// Create user_profiles table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		user_id UUID PRIMARY KEY,

		-- Profile Info
		bio TEXT,
		location VARCHAR(100),
		website VARCHAR(255),

		-- Social Stats
		karma INTEGER DEFAULT 0,
		followers_count INTEGER DEFAULT 0,
		following_count INTEGER DEFAULT 0,

		-- Timestamps
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),

		-- Foreign Key to users_cache (same DB)
		CONSTRAINT fk_user_profiles_user
			FOREIGN KEY (user_id)
			REFERENCES users_cache(id)
			ON DELETE CASCADE
	);
	`

	if err := db.Exec(createTableSQL).Error; err != nil {
		log.Fatalf("❌ Failed to create user_profiles table: %v", err)
	}

	log.Println("✅ user_profiles table created")

	// Create indexes
	log.Println("📊 Creating indexes...")

	indexSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_user_profiles_karma ON user_profiles(karma DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_user_profiles_updated_at ON user_profiles(updated_at)`,
	}

	for _, sql := range indexSQL {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Failed to create index: %v", err)
		}
	}

	log.Println("✅ Indexes created")

	// Auto-create user_profiles for all users_cache entries
	log.Println("📊 Creating default user_profiles for existing users...")

	insertDefaultSQL := `
	INSERT INTO user_profiles (user_id, created_at, updated_at)
	SELECT id, created_at, updated_at
	FROM users_cache
	WHERE NOT EXISTS (
		SELECT 1 FROM user_profiles WHERE user_profiles.user_id = users_cache.id
	)
	`

	result := db.Exec(insertDefaultSQL)
	if result.Error != nil {
		log.Fatalf("❌ Failed to create default profiles: %v", result.Error)
	}

	log.Printf("✅ Created %d default user profiles", result.RowsAffected)

	// Check if old users table exists
	var tableExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists).Error
	if err != nil {
		log.Printf("⚠️  Could not check for old users table: %v", err)
	}

	if tableExists {
		log.Println("📊 Migrating data from old users table...")

		migrateDataSQL := `
		INSERT INTO user_profiles (user_id, bio, location, website, karma, followers_count, following_count, created_at, updated_at)
		SELECT id, bio, location, website, karma, followers_count, following_count, created_at, updated_at
		FROM users
		WHERE EXISTS (SELECT 1 FROM users_cache WHERE users_cache.id = users.id)
		ON CONFLICT (user_id) DO UPDATE SET
			bio = EXCLUDED.bio,
			location = EXCLUDED.location,
			website = EXCLUDED.website,
			karma = EXCLUDED.karma,
			followers_count = EXCLUDED.followers_count,
			following_count = EXCLUDED.following_count,
			updated_at = EXCLUDED.updated_at
		`

		result := db.Exec(migrateDataSQL)
		if result.Error != nil {
			log.Printf("⚠️  Failed to migrate data from old users table: %v", result.Error)
		} else {
			log.Printf("✅ Migrated %d user profiles from old users table", result.RowsAffected)
		}
	} else {
		log.Println("ℹ️  No old users table found, skipping data migration")
	}

	// Summary
	var count int64
	db.Raw("SELECT COUNT(*) FROM user_profiles").Scan(&count)

	log.Println("\n==================================================")
	log.Println("✅ Migration completed successfully!")
	log.Printf("📊 Total user_profiles: %d", count)
	log.Println("==================================================")

	os.Exit(0)
}
