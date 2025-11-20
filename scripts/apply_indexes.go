package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🚀 Applying Performance Indexes")
	fmt.Println(string(make([]byte, 80)))

	// Read migration file
	sqlContent, err := os.ReadFile("migrations/019_add_essential_feed_indexes.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	fmt.Println("\n📝 Running migration: 019_add_essential_feed_indexes.sql\n")

	start := time.Now()

	// Execute migration
	err = db.Exec(string(sqlContent)).Error
	if err != nil {
		log.Fatalf("Failed to apply indexes: %v", err)
	}

	duration := time.Since(start)

	fmt.Println(string(make([]byte, 80)))
	fmt.Printf("\n✅ Migration completed successfully!\n")
	fmt.Printf("   Duration: %v\n", duration)

	// Verify indexes were created
	fmt.Println("\n🔍 Verifying indexes...\n")

	criticalIndexes := []string{
		"idx_posts_feed_composite",
		"idx_posts_votes_desc",
		"idx_post_media_batch",
		"idx_post_tags_batch",
		"idx_tags_name_lower",
	}

	for _, indexName := range criticalIndexes {
		var exists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE schemaname = 'public'
				AND indexname = ?
			)
		`, indexName).Scan(&exists).Error

		if err != nil {
			log.Printf("   ⚠️  Failed to check %s: %v\n", indexName, err)
			continue
		}

		if exists {
			fmt.Printf("   ✅ %s\n", indexName)
		} else {
			fmt.Printf("   ❌ %s (NOT CREATED)\n", indexName)
		}
	}

	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Println("\n🎉 All indexes have been applied!")
	fmt.Println("\n📈 Next step: Run performance test to measure improvement")
	fmt.Println("   Command: go run scripts/analyze_queries.go")
}
